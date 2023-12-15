package main

import (
	"fmt"
	"log"
	"net/http"

	websocket "ninetynine/websocket"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"github.com/gorilla/mux"
)

var FirestoreClient *firestore.Client
var AuthClient *auth.Client

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

func main() {
	// firebase setup
	InitializeFirebase()

	router := mux.NewRouter()

	// request handlers
	router.HandleFunc("/register", registerHandler)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/createroom", createroomHandler)
	router.HandleFunc("/joinroom", joinroomHandler)

	// websocket handlers
	router.HandleFunc("/ws/{roomId}", websocketHandler)

	port := ":8080" // Port number to listen on

	// Create a new HTTP server with default handler
	server := &http.Server{
		Addr:    port,
		Handler: router, // Using default handler (nil)
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

	pool.Register <- client
	client.Read()
}
