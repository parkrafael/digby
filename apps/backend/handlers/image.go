package handlers

import (
	"net/http"

	"backend/db"
)

func GetImage(w http.ResponseWriter, r *http.Request) {
	imageID := r.PathValue("image_id")
	if imageID == "" {
		http.Error(w, "missing image_id", http.StatusBadRequest)
		return
	}

	var fileName, agentID, fileType string
	err := db.DB.QueryRow(
		r.Context(),
		"SELECT file_name, agent_id, file_type FROM images WHERE image_id = $1",
		imageID,
	).Scan(&fileName, &agentID, &fileType)
	if err != nil {
		http.Error(w, "image not found", http.StatusNotFound)
		return
	}

	resp, err := SendTunnelRequest(agentID, tunnelRequest{
		RequestID: imageID,
		FileName:  fileName,
	})
	if err != nil {
		http.Error(w, "agent offline", http.StatusServiceUnavailable)
		return
	}

	if resp.Error != "" {
		http.Error(w, resp.Error, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", fileType)
	w.Write(resp.Bytes)
}