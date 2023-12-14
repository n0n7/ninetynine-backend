package main

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func InitializeFirebase() {
	// Fetch the service account key JSON file path from environment variable or specify it directly
	opt := option.WithCredentialsFile("serviceAccountKey.json")

	// Initialize the app with the service account key
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
	}

	// Access Firestore client
	FirestoreClient, err = app.Firestore(context.Background())
	if err != nil {
		log.Fatalf("Error initializing Firestore client: %v", err)
	}

	AuthClient, err = app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Error initializing Auth client: %v", err)
	}
	fmt.Println("Firebase initialized successfully!")
}
