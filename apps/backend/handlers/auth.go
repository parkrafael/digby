package handlers

import (
	"backend/db"
	"backend/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/resend/resend-go/v3"
	"net/http"
	"os"
	"time"
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
		"INSERT INTO users (email, agent_id) VALUES ($1, $2)",
		user.Email,
		user.AgentID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		// if err is not due to existing email (pk) or agent_id (unique)
		if errors.As(err, &pgErr) && pgErr.Code != "23505" {
			http.Error(w, "database error: "+pgErr.Code, http.StatusInternalServerError)
			return
			// if err is not a postgres error
		} else if !errors.As(err, &pgErr) {
			http.Error(w, "database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// generate and store magic link token
	token := uuid.New().String()
	expiresAt := time.Now().Add(time.Minute * 15)
	_, err = db.DB.Exec(
		r.Context(),
		"INSERT INTO magic_links (token, email, expires_at) VALUES ($1, $2, $3)",
		token,
		user.Email,
		expiresAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			http.Error(w, "database error: "+pgErr.Code, http.StatusInternalServerError)
			return
		}
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
        <a href="http://localhost:8080/auth/verify?token=%s">Sign in to Digby</a>
        <p>If you didn't request this email, you can safely ignore it.</p>
    `, token),
	}
	_, err = client.Emails.Send(&params)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to send email: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
