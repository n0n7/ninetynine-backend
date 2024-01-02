package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"ninetynine/firebase"
	Handler "ninetynine/handler"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
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
	router.HandleFunc("/getroom", Handler.GetRoomHandler)

	// websocket handlers
	router.HandleFunc("/ws/{roomId}", Handler.WebsocketHandler)

	// read PORT from .env file
	port := ":" + getEnv("PORT")

	// setup CORS
	corsHandler := cors.Default()
	handler := corsHandler.Handler(router)

	// Create a new HTTP server with default handler
	server := &http.Server{
		Addr:    port,
		Handler: handler, // Using default handler (nil)
	}

	// Start the HTTP server
	fmt.Printf("Server is running on https://localhost%s\n", port)
	err := server.ListenAndServeTLS("fullchain.pem", "privatekey.pem")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key string) string {
	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
