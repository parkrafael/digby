package models

import "time"

type User struct {
	Email   string `json:"email"`
	AgentID string `json:"agent_id"`
}

type Image struct {
	ImageID       string    `json:"image_id"`
	AgentID       string    `json:"agent_id"`
	FileName      string    `json:"file_name"`
	FileType      string    `json:"file_type"`
	FileTimestamp time.Time `json:"file_timestamp"`
	Tags          []string  `json:"tags"`
}
