package handlers

import (
	"errors"
	"net/http"

	"backend/db"

	"github.com/jackc/pgx/v5/pgconn"
)

func IsAgentRegistered(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		http.Error(w, "missing agent_id", http.StatusBadRequest)
		return
	}

	var exists bool
	err := db.DB.QueryRow(
		r.Context(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE agent_id = $1)",
		agentID,
	).Scan(&exists)
	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)
		if ok {
			http.Error(w, "database error: "+pgErr.Code, http.StatusInternalServerError)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "agent not registered", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}