package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"backend/db"
	"backend/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/resend/resend-go/v3"
)

const (
	magicLinkExpiry = 15 * time.Minute
	jwtExpiry       = 24 * time.Hour
)

func SendMagicLink(w http.ResponseWriter, r *http.Request) {
	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// create new user
	_, err = db.DB.Exec(
		r.Context(),
		"INSERT INTO users (email, agent_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		user.Email,
		user.AgentID,
	)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			http.Error(w, "database error: "+pgErr.Code, http.StatusInternalServerError)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// generate and store magic link token
	token := uuid.New().String()
	expiresAt := time.Now().UTC().Add(magicLinkExpiry)
	_, err = db.DB.Exec(
		r.Context(),
		"INSERT INTO magic_links (token, email, expires_at) VALUES ($1, $2, $3)",
		token,
		user.Email,
		expiresAt,
	)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			http.Error(w, "database error: "+pgErr.Code, http.StatusInternalServerError)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// send email
	client := resend.NewClient(os.Getenv("RESEND_API_KEY"))
	params := resend.SendEmailRequest{
		From:    "noreply@digbyapp.xyz",
		To:      []string{user.Email},
		Subject: "Your Digby login link",
		Html: fmt.Sprintf(`
        <p>Hello,</p>
        <p>Click the link below to sign in to Digby. Link expires in 15 minutes.</p>
        <a href="%s/auth/verify?token=%s">Sign in to Digby</a>
        <p>If you didn't request this email, you can safely ignore it.</p>
    `, os.Getenv("FRONTEND_URL"), token),
	}
	_, err = client.Emails.Send(&params)
	if err != nil {
		http.Error(w, "failed to send email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func VerifyToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	var magicLink struct {
		Token     string    `json:"token"`
		Email     string    `json:"email"`
		ExpiresAt time.Time `json:"expires_at"`
		CreatedAt time.Time `json:"created_at"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// retrieve row if token exists
	err = db.DB.QueryRow(
		r.Context(),
		"SELECT token, email, expires_at, created_at FROM magic_links WHERE token = $1",
		body.Token,
	).Scan(&magicLink.Token, &magicLink.Email, &magicLink.ExpiresAt, &magicLink.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// validate token hasn't expired
	if time.Now().UTC().After(magicLink.ExpiresAt) {
		http.Error(w, "expired token", http.StatusBadRequest)
		return
	}

	// delete token from database
	_, err = db.DB.Exec(
		r.Context(),
		"DELETE FROM magic_links WHERE token = $1",
		body.Token,
	)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			http.Error(w, "database error: "+pgErr.Code, http.StatusInternalServerError)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// generate jwt
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": magicLink.Email,
		"exp":   time.Now().UTC().Add(jwtExpiry).Unix(),
	}).SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
