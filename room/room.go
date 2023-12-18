package room

import (
	"context"
	"math/rand"
	Firebase "ninetynine/firebase"
	"strconv"
	"time"

	"google.golang.org/api/iterator"
)

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

func CreateRoom(userId string) (Room, error) {
	roomId, err := generateRoomId()
	if err != nil {
		return Room{}, err
	}

	// create new room
	newRoom := Room{
		RoomID:       roomId,
		CreatedAt:    time.Now().Unix(),
		OwnerID:      userId,
		MaxCapacity:  8,
		MaxSpectator: 16,
		Status:       "waiting",
		Players:      []string{},
		Spectators:   []string{},
	}

	_, err = Firebase.FirestoreClient.Collection("rooms").Doc(roomId).Set(context.Background(), newRoom)
	if err != nil {
		return Room{}, err
	}

	return newRoom, nil

}

func JoinRoom(userId string, roomId string) (Room, error, string) {
	// find room in firestore
	roomData, err := findRoom(roomId)
	if err != nil {
		return Room{}, err, "Error finding room"
	}

	// check if room exists
	if roomData.RoomID == "" {
		return Room{}, nil, "Room does not exist"
	}

	// check if user is already in room
	for _, player := range roomData.Players {
		if player == userId {
			return roomData, nil, ""
		}
	}

	// check if room is full
	playerCount := len(roomData.Players)

	if playerCount >= roomData.MaxCapacity {
		return Room{}, nil, "Room is full"
	}

	// check room status
	if roomData.Status != "waiting" {
		return Room{}, nil, "Room is not open"
	}

	// add player to room
	roomData.Players = append(roomData.Players, userId)

	// if the room is full, change status to "full"
	if playerCount+1 == roomData.MaxCapacity {
		roomData.Status = "full"
	}

	// update room in firestore
	_, err = updateRoom(roomId, roomData)
	if err != nil {
		return Room{}, err, "Error updating room"
	}

	return roomData, nil, ""
}

func findRoom(roomId string) (Room, error) {
	// find room in firestore
	docRef := Firebase.FirestoreClient.Collection("rooms").Doc(roomId)
	docSnap, err := docRef.Get(context.Background())
	if err == iterator.Done {
		return Room{}, nil
	}

	if err != nil {
		return Room{}, err
	}

	data := docSnap.Data()

	roomData := Room{
		RoomID:       roomId,
		CreatedAt:    data["createdAt"].(int64),
		OwnerID:      data["ownerId"].(string),
		MaxCapacity:  data["maxCapacity"].(int),
		MaxSpectator: data["maxSpectator"].(int),
		Status:       data["status"].(string),
		Players:      data["players"].([]string),
		Spectators:   data["spectators"].([]string),
	}

	return roomData, nil
}

func updateRoom(roomId string, roomData Room) (Room, error) {
	// update room in firestore

	docRef := Firebase.FirestoreClient.Collection("rooms").Doc(roomId)
	_, err := docRef.Set(context.Background(), roomData)
	if err != nil {
		return Room{}, err
	}

	return roomData, nil
}

func generateRoomId() (string, error) {
	for {
		// Generate a random 12-digit room ID
		rand.NewSource(time.Now().UnixNano())
		roomID := strconv.Itoa(rand.Intn(1e12))

		// check if room exists
		docRef := Firebase.FirestoreClient.Collection("rooms").Doc(roomID)
		_, err := docRef.Get(context.Background())
		if err == iterator.Done {
			return roomID, nil
		}

		if err != nil {
			return "", err
		}
	}
}
