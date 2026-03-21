package main

import (
	"encoding/json"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /users", getUsers)
	mux.HandleFunc("POST /users", createUser)
	mux.HandleFunc("GET /users/{id}", getUserByID)

	http.ListenAndServe(":8080", mux)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "get all users"})
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	json.NewDecoder(r.Body).Decode(&body)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body)
}

func getUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // grab path param

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}