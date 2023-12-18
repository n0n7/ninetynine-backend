package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"ninetynine/firebase"
	Firebase "ninetynine/firebase"
	websocket "ninetynine/websocket"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"google.golang.org/api/iterator"
)

func main() {
	// firebase setup
	firebase.InitializeFirebase()

	router := mux.NewRouter()

	// request handlers
	router.HandleFunc("/register", registerHandler)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/createroom", createroomHandler)
	router.HandleFunc("/joinroom", joinroomHandler)

	// websocket handlers
	router.HandleFunc("/ws/{roomId}", websocketHandler)

	port := ":8080" // Port number to listen on

	// setup CORS
	corsHandler := cors.Default()
	handler := corsHandler.Handler(router)

	// Create a new HTTP server with default handler
	server := &http.Server{
		Addr:    port,
		Handler: handler, // Using default handler (nil)
	}

	// Start the HTTP server
	fmt.Printf("Server is running on http://localhost%s\n", port)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

var Pools = make(map[string]*websocket.Pool)

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	roomId := mux.Vars(r)["roomId"]

	// check if room exist in firestore
	docRef := Firebase.FirestoreClient.Collection("rooms").Doc(roomId)
	_, err := docRef.Get(context.Background())
	if err == iterator.Done {
		fmt.Println("Room", roomId, "does not exist")
		handleRequestError(w, "room does not exist", http.StatusBadRequest)
		return
	}

	if err != nil {
		fmt.Println("Error getting document", err)
		handleRequestError(w, "error getting document", http.StatusInternalServerError)
		return
	}

	if _, exist := Pools[roomId]; !exist {
		fmt.Println("Creating new pool for room", roomId)
		Pools[roomId] = websocket.NewPool(roomId)
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

	go ManageRoom(pool.RoomId)

	pool.Register <- client
	client.Read()
}

func ManageRoom(roomId string) {
	defer func() {
		delete(websocket.GameStateChange, roomId)
	}()

	for {
		select {
		case gameStatus := <-websocket.GameStateChange[roomId]:
			// get roomData from firestore
			docRef := Firebase.FirestoreClient.Collection("rooms").Doc(roomId)
			docSnap, err := docRef.Get(context.Background())
			if err != nil {
				fmt.Println("Error getting document", err)
				break
			}

			fmt.Println("game status update", gameStatus)

			roomData := docSnap.Data()

			// edit status
			roomData["status"] = gameStatus

			// update room in firestore
			_, err = docRef.Set(context.Background(), roomData)
			if err != nil {
				fmt.Println("Error updating document", err)
				break
			}

			if gameStatus == "ended" {
				return
			}

		}
	}

}
