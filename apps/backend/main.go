package main

import (
	"backend/db"
	"backend/handlers"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	err = db.Connect()
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	// Authentication routes
	mux.HandleFunc("POST /auth/magic-link", handlers.SendMagicLink)

	http.ListenAndServe(":8080", mux)
}
