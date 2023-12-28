package room

import (
	"context"

	Firebase "ninetynine/firebase"
)

func GetRoom(roomId string) (map[string]interface{}, error) {
	docRef := Firebase.FirestoreClient.Collection("rooms").Doc(roomId)
	docSnap, err := docRef.Get(context.Background())
	if err != nil {
		return nil, err
	}

	roomData := docSnap.Data()

	return roomData, nil
}
