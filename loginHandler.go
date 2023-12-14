package main

import (
	"context"
	"encoding/json"
	"net/http"

	"google.golang.org/api/iterator"
)

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
	query := FirestoreClient.Collection("users").Where("email", "==", loginData.Email).Limit(1)
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
