package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
)

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
