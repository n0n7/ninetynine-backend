package main

import (
	"context"
	"encoding/json"
	"net/http"

	"google.golang.org/api/iterator"
)

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
	docRef := FirestoreClient.Collection("rooms").Doc(roomId)
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
