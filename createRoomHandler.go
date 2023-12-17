package main

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/api/iterator"
)

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
	docRef := FirestoreClient.Collection("users").Doc(data["userId"].(string))
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
		Status:       "waiting",
		Players:      []string{},
		Spectators:   []string{},
	}

	// add room to firestore
	_, err = FirestoreClient.Collection("rooms").Doc(roomID).Set(context.Background(), newRoom)
	if err != nil {
		handleRequestError(w, "Error creating room", http.StatusInternalServerError)
		return
	}

	// write response
	w.WriteHeader(http.StatusOK)
	responseJSON, _ := json.Marshal(newRoom)
	w.Write(responseJSON)

}
