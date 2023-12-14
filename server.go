package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"crypto/sha256"
	"encoding/hex"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var firestoreClient *firestore.Client
var authClient *auth.Client

type RegisterData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

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
	initializeFirebase()

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

func initializeFirebase() {
	// Fetch the service account key JSON file path from environment variable or specify it directly
	opt := option.WithCredentialsFile("serviceAccountKey.json")

	// Initialize the app with the service account key
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
	}

	// Access Firestore client
	firestoreClient, err = app.Firestore(context.Background())
	if err != nil {
		log.Fatalf("Error initializing Firestore client: %v", err)
	}

	authClient, err = app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Error initializing Auth client: %v", err)
	}
	fmt.Println("Firebase initialized successfully!")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// only allow POST requests
	if r.Method != http.MethodPost {
		handleRequestError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// check request body format
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		handleRequestError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	requiredFields := []string{"username", "password", "email"}
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			handleRequestError(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	var newUser RegisterData
	newUser.Username = data["username"].(string)
	newUser.Password = data["password"].(string)
	newUser.Email = data["email"].(string)

	// check if email is valid
	if !IsValidEmail(newUser.Email) {
		handleRequestError(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// check if email already exists
	_, err = authClient.GetUserByEmail(context.Background(), newUser.Email)
	if err == nil {
		handleRequestError(w, "User with this email already exists", http.StatusBadRequest)
		return
	}

	// check if username already exists in firestore
	query := firestoreClient.Collection("users").Where("username", "==", newUser.Username).Limit(1)
	docSnap, err := query.Documents(context.Background()).Next()
	if err != nil && err != iterator.Done {
		handleRequestError(w, "Error checking username", http.StatusInternalServerError)
		return
	}
	if docSnap != nil {
		handleRequestError(w, "User with this username already exists", http.StatusBadRequest)
		return
	}

	// create user in auth
	params := (&auth.UserToCreate{}).
		Email(newUser.Email).
		Password(newUser.Password)

	userRecord, err := authClient.CreateUser(context.Background(), params)
	if err != nil {
		handleRequestError(w, "Error creating user", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	// create user in firestore
	docRef := firestoreClient.Collection("users").Doc(userRecord.UID)

	hashedPassword := hashedPassword(newUser.Password)

	userData := map[string]interface{}{
		"username":  newUser.Username,
		"email":     newUser.Email,
		"userId":    userRecord.UID,
		"createdAt": time.Now().Unix(), // Current timestamp (UNIX time)
		"password":  hashedPassword,
		"gamestat": map[string]interface{}{
			"playCount": 0,
		},
	}

	_, err = docRef.Set(context.Background(), userData)
	if err != nil {
		handleRequestError(w, "Error registering user", http.StatusInternalServerError)
		return
	}

	delete(userData, "password")

	// write response
	w.WriteHeader(http.StatusOK)
	responseJSON, _ := json.Marshal(userData)
	w.Write(responseJSON)

}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// only allow POST requests
	if r.Method != http.MethodPost {
		handleRequestError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// check request body format
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		handleRequestError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	requiredFields := []string{"email", "password"}
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			handleRequestError(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	var loginData LoginData
	loginData.Email = data["email"].(string)
	loginData.Password = data["password"].(string)

	hashedPassword := hashedPassword(loginData.Password)

	// authenticate user in firestore
	query := firestoreClient.Collection("users").Where("email", "==", loginData.Email).Limit(1)
	docSnap, err := query.Documents(context.Background()).Next()
	if err == iterator.Done {
		handleRequestError(w, "User with this email does not exist", http.StatusBadRequest)
		return
	}

	if err != nil {
		handleRequestError(w, "Error checking email", http.StatusInternalServerError)
		return
	}

	userData := docSnap.Data()
	if userData["password"] != hashedPassword {
		handleRequestError(w, "Incorrect password", http.StatusBadRequest)
		return
	}

	// write response
	w.WriteHeader(http.StatusOK)
	responseJSON, _ := json.Marshal(userData)
	w.Write(responseJSON)

}

func createroomHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// check method
	if r.Method != http.MethodPost {
		handleRequestError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// check user is logged in
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		handleRequestError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if _, exists := data["userId"]; !exists {
		handleRequestError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// find user in firestore
	docRef := firestoreClient.Collection("users").Doc(data["userId"].(string))
	_, err = docRef.Get(context.Background())
	if err == iterator.Done {
		handleRequestError(w, "User does not exist", http.StatusBadRequest)
		return
	}
	if err != nil {
		handleRequestError(w, "Error finding user", http.StatusInternalServerError)
		return
	}

	// Generate a random 12-digit room ID
	rand.NewSource(time.Now().UnixNano())
	roomID := strconv.Itoa(rand.Intn(1e12))

	// create new room
	newRoom := Room{
		RoomID:       roomID,
		CreatedAt:    time.Now().Unix(),
		OwnerID:      data["userId"].(string),
		MaxCapacity:  8,
		MaxSpectator: 16,
		Status:       "open",
		Players:      []string{},
		Spectators:   []string{},
	}

	// add room to firestore
	_, err = firestoreClient.Collection("rooms").Doc(roomID).Set(context.Background(), newRoom)
	if err != nil {
		handleRequestError(w, "Error creating room", http.StatusInternalServerError)
		return
	}

	// write response
	w.WriteHeader(http.StatusOK)
	responseJSON, _ := json.Marshal(newRoom)
	w.Write(responseJSON)

}

func joinroomHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// check method
	if r.Method != http.MethodPost {
		handleRequestError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// check user is logged in
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		handleRequestError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if _, exists := data["userId"]; !exists {
		handleRequestError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// check roomId exists
	if _, exists := data["roomId"]; !exists {
		handleRequestError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	roomId := data["roomId"].(string)

	// find room in firestore
	docRef := firestoreClient.Collection("rooms").Doc(roomId)
	docSnap, err := docRef.Get(context.Background())
	if err == iterator.Done {
		handleRequestError(w, "Room does not exist", http.StatusBadRequest)
		return
	}

	if err != nil {
		handleRequestError(w, "Error finding room", http.StatusInternalServerError)
		return
	}

	// check if room is full
	roomData := docSnap.Data()
	playerCount := len(roomData["players"].([]interface{}))

	if int64(playerCount) >= roomData["maxCapacity"].(int64) {
		handleRequestError(w, "Room is full", http.StatusBadRequest)
		return
	}

	// check room status
	if roomData["status"].(string) != "open" {
		handleRequestError(w, "Room is not open", http.StatusBadRequest)
		return
	}

	// add player to room
	roomData["players"] = append(roomData["players"].([]string), data["userId"].(string))

	// if the room is full, change status to "full"
	if int64(playerCount+1) == roomData["maxCapacity"].(int64) {
		roomData["status"] = "full"
	}

	// update room in firestore
	_, err = docRef.Set(context.Background(), roomData)
	if err != nil {
		handleRequestError(w, "Error joining room", http.StatusInternalServerError)
		return
	}

	// write response
	w.WriteHeader(http.StatusOK)
	responseJSON, _ := json.Marshal(roomData)
	w.Write(responseJSON)

}

func handleRequestError(w http.ResponseWriter, errMessage string, statusCode int) {
	responseData := map[string]interface{}{
		"error": errMessage,
	}

	responseJSON, _ := json.Marshal(responseData)
	w.WriteHeader(statusCode)
	w.Write(responseJSON)
}

func hashedPassword(password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	hashed := hasher.Sum(nil)
	return hex.EncodeToString(hashed)
}

func IsValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(emailRegex, email)
	return match
}
