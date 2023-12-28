package handler

import (
	"context"
	"fmt"
	"net/http"

	Firebase "ninetynine/firebase"
	Room "ninetynine/room"
	websocket "ninetynine/websocket"

	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

var Pools = make(map[string]*websocket.Pool)

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	roomId := mux.Vars(r)["roomId"]

	// check if room exist in firestore
	docRef := Firebase.FirestoreClient.Collection("rooms").Doc(roomId)
	_, err := docRef.Get(context.Background())
	if err == iterator.Done {
		fmt.Println("Room", roomId, "does not exist")
		return
	}

	if err != nil {
		fmt.Println("Error getting document", err)
		return
	}

	docSnap, err := docRef.Get(context.Background())
	if err != nil {
		fmt.Println("Error getting document", err)
		return
	}

	roomData := docSnap.Data()

	if _, exist := Pools[roomId]; !exist {
		fmt.Println("Creating new pool for room", roomId)
		Pools[roomId] = websocket.NewPool(roomId, roomData["ownerId"].(string))
		go Pools[roomId].Start()
	}

	fmt.Println("WebSocket Endpoint Hit for room", roomId)
	serveWs(Pools[roomId], w, r)

}

func serveWs(pool *websocket.Pool, w http.ResponseWriter, r *http.Request) {
	defer func() {
		if len(pool.Clients) == 0 {
			delete(Pools, pool.RoomId)
			fmt.Println("Deleting room", pool.RoomId)
		}
	}()
	conn, err := websocket.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
	}

	client := &websocket.Client{
		Conn: conn,
		Pool: pool,
	}

	go Room.ManageRoom(pool.RoomId)

	pool.Register <- client
	client.Read()
}
