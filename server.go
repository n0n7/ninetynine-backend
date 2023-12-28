package main

import (
	"fmt"
	"log"
	"net/http"

	"ninetynine/firebase"
	Handler "ninetynine/handler"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// firebase setup
	firebase.InitializeFirebase()

	router := mux.NewRouter()

	// request handlers
	router.HandleFunc("/register", Handler.RegisterHandler)
	router.HandleFunc("/login", Handler.LoginHandler)
	router.HandleFunc("/createroom", Handler.CreateroomHandler)
	router.HandleFunc("/joinroom", Handler.JoinroomHandler)

	// websocket handlers
	router.HandleFunc("/ws/{roomId}", Handler.WebsocketHandler)

	port := ":8080" // Port number to listen on

	// setup CORS
	corsHandler := cors.Default()
	handler := corsHandler.Handler(router)

	// Create a new HTTP server with default handler
	server := &http.Server{
		Addr:    port,
		Handler: handler, // Using default handler (nil)
	}

	// Start the HTTP server
	fmt.Printf("Server is running on http://localhost%s\n", port)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
