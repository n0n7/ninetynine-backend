package main

import (
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
)

var FirestoreClient *firestore.Client
var AuthClient *auth.Client

type Room struct {
	RoomID       string   `json:"roomId"`
	CreatedAt    int64    `json:"createTime"`
	OwnerID      string   `json:"ownerId"`
	MaxCapacity  int      `json:"maxCapacity"`
	MaxSpectator int      `json:"maxSpectator"`
	Status       string   `json:"status"`
	Players      []string `json:"players"`
	Spectators   []string `json:"spectators"`
}

func main() {
	// firebase setup
	InitializeFirebase()

	// request handlers
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/createroom", createroomHandler)
	http.HandleFunc("/joinroom", joinroomHandler)

	port := ":8080" // Port number to listen on

	// Create a new HTTP server with default handler
	server := &http.Server{
		Addr:    port,
		Handler: nil, // Using default handler (nil)
	}

	// Start the HTTP server
	fmt.Printf("Server is running on http://localhost%s\n", port)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
