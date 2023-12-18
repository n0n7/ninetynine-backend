package firebase

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	fb "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var FirestoreClient *firestore.Client
var AuthClient *auth.Client

func InitializeFirebase() {
	// Fetch the service account key JSON file path from environment variable or specify it directly
	opt := option.WithCredentialsFile("serviceAccountKey.json")

	// Initialize the app with the service account key
	app, err := fb.NewApp(context.Background(), nil, opt)
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
