package handler

import (
	"encoding/json"
	"net/http"

	Auth "ninetynine/auth"
	Room "ninetynine/room"
)

func CreateroomHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// check method
	if r.Method != http.MethodPost {
		requestErrorHandler(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// check valid format
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		requestErrorHandler(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	requiredFields := []string{"userId"}
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			requestErrorHandler(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	// check userId
	userId := data["userId"].(string)
	isValid, err := Auth.IsValidUserId(userId)

	if err != nil {
		requestErrorHandler(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !isValid {
		requestErrorHandler(w, "Invalid user", http.StatusBadRequest)
		return
	}

	// create new room
	newRoom, err := Room.CreateRoom(userId)
	if err != nil {
		requestErrorHandler(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// write response
	w.WriteHeader(http.StatusOK)
	responseJSON, _ := json.Marshal(newRoom)
	w.Write(responseJSON)

}
