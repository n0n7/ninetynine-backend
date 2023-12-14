package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"firebase.google.com/go/auth"
	"google.golang.org/api/iterator"
)

type RegisterData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
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
	_, err = AuthClient.GetUserByEmail(context.Background(), newUser.Email)
	if err == nil {
		handleRequestError(w, "User with this email already exists", http.StatusBadRequest)
		return
	}

	// check if username already exists in firestore
	query := FirestoreClient.Collection("users").Where("username", "==", newUser.Username).Limit(1)
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

	userRecord, err := AuthClient.CreateUser(context.Background(), params)
	if err != nil {
		handleRequestError(w, "Error creating user", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	// create user in firestore
	docRef := FirestoreClient.Collection("users").Doc(userRecord.UID)

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
