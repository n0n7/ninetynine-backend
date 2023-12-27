package room

import (
	"context"
	"fmt"
	"math/rand"
	Firebase "ninetynine/firebase"
	"reflect"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Room struct {
	RoomID       string   `json:"roomId"`
	CreatedAt    int64    `json:"createdAt"`
	OwnerID      string   `json:"ownerId"`
	MaxCapacity  int      `json:"maxCapacity"`
	MaxSpectator int      `json:"maxSpectator"`
	Status       string   `json:"status"`
	Players      []string `json:"players"`
	Spectators   []string `json:"spectators"`
}

func RoomToMap(room Room) (map[string]interface{}, error) {
	roomMap := make(map[string]interface{})

	reflectValue := reflect.ValueOf(room)
	reflectType := reflect.TypeOf(room)

	for i := 0; i < reflectValue.NumField(); i++ {
		field := reflectValue.Field(i)
		fieldName := reflectType.Field(i).Tag.Get("json")

		// Ignore fields with empty tag or unsupported types
		if fieldName == "" || field.Kind() == reflect.Invalid {
			continue
		}

		roomMap[fieldName] = field.Interface()
	}

	return roomMap, nil
}

func CreateRoom(userId string) (Room, error) {
	roomId, err := generateRoomId()
	if err != nil {
		fmt.Println(err)
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
		Players:      []string{userId},
		Spectators:   []string{},
	}

	jsonData, _ := RoomToMap(newRoom)
	fmt.Println(jsonData)

	_, err = Firebase.FirestoreClient.Collection("rooms").Doc(roomId).Set(context.Background(), jsonData)
	if err != nil {
		fmt.Println(err)
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
	if status.Code(err) == codes.NotFound {
		return Room{}, nil
	}

	if err != nil {
		return Room{}, err
	}

	data := docSnap.Data()

	fmt.Println(data)

	roomData := Room{
		RoomID:       roomId,
		CreatedAt:    data["createdAt"].(int64),
		OwnerID:      data["ownerId"].(string),
		MaxCapacity:  int(data["maxCapacity"].(int64)),
		MaxSpectator: int(data["maxSpectator"].(int64)),
		Status:       data["status"].(string),
		Players:      toStringSlice(data["players"].([]interface{})),
		Spectators:   toStringSlice(data["spectators"].([]interface{})),
	}

	fmt.Println(roomData)

	return roomData, nil

}

// Helper function to convert []interface{} to []string
func toStringSlice(slice []interface{}) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = v.(string)
	}
	return result
}

func updateRoom(roomId string, roomData Room) (Room, error) {
	// update room in firestore

	docRef := Firebase.FirestoreClient.Collection("rooms").Doc(roomId)
	jsonData, _ := RoomToMap(roomData)
	_, err := docRef.Set(context.Background(), jsonData)
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

		// fill zeroes if roomID is less than 12 digits
		for len(roomID) < 12 {
			roomID = "0" + roomID
		}

		// check if room exists
		docRef := Firebase.FirestoreClient.Collection("rooms").Doc(roomID)
		_, err := docRef.Get(context.Background())
		if status.Code(err) == codes.NotFound {
			return roomID, nil
		}

		if err != nil {
			return "", err
		}
	}
}
