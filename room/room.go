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
