package room

import (
	"context"
	"fmt"
	Firebase "ninetynine/firebase"
)

type FirebaseUpdateData struct {
	Field string
	Value interface{}
}

var FirebaseUpdateChannel = make(map[string]chan FirebaseUpdateData)

func ManageRoom(roomId string) {
	defer func() {
		delete(FirebaseUpdateChannel, roomId)
	}()

	FirebaseUpdateChannel[roomId] = make(chan FirebaseUpdateData)

	for {
		select {
		case updateData := <-FirebaseUpdateChannel[roomId]:

			docRef := Firebase.FirestoreClient.Collection("rooms").Doc(roomId)
			docSnap, err := docRef.Get(context.Background())
			if err != nil {
				fmt.Println("Error getting document", err)
				break
			}

			roomData := docSnap.Data()

			roomData[updateData.Field] = updateData.Value

			// update room in firestore
			_, err = docRef.Set(context.Background(), roomData)
			if err != nil {
				fmt.Println("Error updating document", err)
				break
			}

			// if game ended
			if updateData.Field == "status" && updateData.Value == "ended" {
				return
			}

			break
		}
	}

}

func PlayerLeft(roomId string, playerId string, isOwner bool) string {
	docRef := Firebase.FirestoreClient.Collection("rooms").Doc(roomId)
	docSnap, err := docRef.Get(context.Background())
	if err != nil {
		fmt.Println("Error getting document", err)
		return ""
	}

	roomData := docSnap.Data()
	players := roomData["players"].([]interface{})

	// remove player from players array
	for i, p := range players {
		if p.(string) == playerId {
			players = append(players[:i], players[i+1:]...)
			break
		}
	}

	fmt.Println("players", players)

	FirebaseUpdateChannel[roomId] <- FirebaseUpdateData{
		Field: "players",
		Value: players,
	}

	if isOwner {
		if len(players) > 0 {
			FirebaseUpdateChannel[roomId] <- FirebaseUpdateData{
				Field: "ownerId",
				Value: players[0].(string),
			}
		}

		return players[0].(string)

	} else {
		return roomData["ownerId"].(string)
	}

}
