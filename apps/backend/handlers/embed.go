package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"backend/db"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pgvector/pgvector-go"
)

func EmbedImage(w http.ResponseWriter, r *http.Request) {
	// keep up to 32 MB in memory
	err := r.ParseMultipartForm(32 * 1024 * 1024)
	if err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// (file, meta, err) -> ignore metadata as it's collected from multipart form
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	// format file_timestamp from string to timestamptz
	fileTimestamp, err := time.Parse(time.RFC3339, r.FormValue("file_timestamp"))
	if err != nil {
		http.Error(w, "invalid file_timestamp", http.StatusBadRequest)
		return
	}

	_, err = db.DB.Exec(
		r.Context(),
		`INSERT INTO images (image_id, agent_id, file_name, file_type, file_timestamp, tags)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		r.FormValue("image_id"),
		r.FormValue("agent_id"),
		r.FormValue("file_name"),
		r.FormValue("file_type"),
		fileTimestamp,
		strings.Split(r.FormValue("tags"), ","),
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

	// create multipart form for clip embedding service
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", r.FormValue("file_name"))
	if err != nil {
		http.Error(w, "failed to create multipart form", http.StatusInternalServerError)
		return
	}
	_, err = part.Write(fileBytes)
	if err != nil {
		http.Error(w, "failed to write file", http.StatusInternalServerError)
		return
	}
	writer.Close()

	// call clip service
	resp, err := http.Post(os.Getenv("CLIP_SERVICE_URL")+"/embed/image", writer.FormDataContentType(), &body)
	if err != nil {
		http.Error(w, "clip service error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "clip service error", http.StatusInternalServerError)
		return
	}

	// store embedding in images table
	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		http.Error(w, "failed to decode embedding", http.StatusInternalServerError)
		return
	}
	_, err = db.DB.Exec(
		r.Context(),
		`UPDATE images SET embedding = $1 WHERE image_id = $2`,
		pgvector.NewVector(result.Embedding),
		r.FormValue("image_id"),
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

	w.WriteHeader(http.StatusOK)
}
