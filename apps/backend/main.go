package main

import (
	"log"
	"net/http"

	"backend/db"
	"backend/handlers"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	err = db.Connect()
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	// Authentication routes
	mux.HandleFunc("POST /auth/magic-link", handlers.SendMagicLink)
	mux.HandleFunc("POST /auth/verify", handlers.VerifyToken)

	// Agent
	mux.HandleFunc("GET /agent/registered", handlers.IsAgentRegistered)

	http.ListenAndServe(":8080", mux)
}
