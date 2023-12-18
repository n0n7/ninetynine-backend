package main

import (
	"encoding/json"
	"net/http"
	"time"

	Auth "ninetynine/auth"
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
	if !Auth.IsValidEmail(newUser.Email) {
		handleRequestError(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// check if email already exists
	isUniqueEmail, err := Auth.CheckUniqueEmail(newUser.Email)
	if err != nil {
		handleRequestError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isUniqueEmail {
		handleRequestError(w, "User with this email already exists", http.StatusBadRequest)
		return
	}

	// check if username already exists in firestore
	isUniqueUsername, err := Auth.CheckUniqueUsername(newUser.Username)
	if err != nil {
		handleRequestError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if !isUniqueUsername {
		handleRequestError(w, "User with this username already exists", http.StatusBadRequest)
		return
	}

	// create user in firestore
	userData := map[string]interface{}{
		"username":  newUser.Username,
		"email":     newUser.Email,
		"createdAt": time.Now().Unix(), // Current timestamp (UNIX time)
		"password":  newUser.Password,  // get hashed in Auth.CreateUser function
		"gamestat": map[string]interface{}{
			"playCount": 0,
		},
	}

	userData, err = Auth.CreateUser(userData)
	if err != nil {
		handleRequestError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// write response
	w.WriteHeader(http.StatusOK)
	responseJSON, _ := json.Marshal(userData)
	w.Write(responseJSON)

}
