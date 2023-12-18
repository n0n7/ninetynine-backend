package websocket

import "fmt"

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan Message
	GameAction chan string
	RoomId     string
	Game       *Game
}

func NewPool(RoomId string) *Pool {
	newGame := NewGame()
	go newGame.Start()
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message),
		GameAction: make(chan string),
		RoomId:     RoomId,
		Game:       newGame,
	}
}

func (pool *Pool) Start() {
	defer func() {
		for client := range pool.Clients {
			client.Conn.Close()
		}
		pool.Game.Stop <- true
	}()

	GameStateChange[pool.RoomId] = make(chan string)

	pool.Game.Pool = pool

	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))
			break
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
			pool.Game.Unregister <- client.ID
			fmt.Println("Size of Connection Pool: ", len(pool.Clients))

			if len(pool.Clients) == 0 {
				return
			}
			break
		case message := <-pool.Broadcast:
			for client := range pool.Clients {
				if err := client.Conn.WriteJSON(message); err != nil {
					fmt.Println(err)
					return
				}
			}
			break
		case actionMessage := <-pool.GameAction:
			pool.BroadCaseGameData(actionMessage)
		}
	}
}

func (pool *Pool) BroadCaseGameData(actionMessage string) {
	fmt.Println("broadcast game data", actionMessage)
	for client := range pool.Clients {
		gameData := pool.Game.GetGameData(client.ID)
		client.Conn.WriteJSON(Message{Error: "", Action: actionMessage, GameData: gameData})
	}
}
