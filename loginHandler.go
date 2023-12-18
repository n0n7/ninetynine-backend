package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	Auth "ninetynine/auth"
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

	// authenticate user in firestore
	isValid, userData, err := Auth.Login(loginData.Email, loginData.Password)
	if err != nil {
		fmt.Println(err)
		handleRequestError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !isValid {
		handleRequestError(w, "email or password is incorrect", http.StatusBadRequest)
		return
	}

	// write response
	w.WriteHeader(http.StatusOK)
	responseJSON, _ := json.Marshal(userData)
	w.Write(responseJSON)

}
