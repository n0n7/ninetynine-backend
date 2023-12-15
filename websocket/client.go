package websocket

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
	Pool *Pool
}

type Message struct {
	Error    string      `json:"error"`
	Action   string      `json:"action"`
	GameData GameMessage `json:"gameData"`
}

type PlayerMessage struct {
	PlayerId        string `json:"playerId"`
	PlayerName      string `json:"playerName"`
	PlayerAvatarURL string `json:"playerAvatarURL"`
	IsOut           bool   `json:"isOut"`
	Status          string `json:"status"`
}

type GameMessage struct {
	Players            []PlayerMessage `json:"players"`
	PlayerCards        []Card          `json:"playerCards"`
	Status             string          `json:"status"`
	CurrentPlayerIndex int             `json:"currentPlayerIndex"`
	CurrentDirection   int             `json:"currentDirection"`
	StackValue         int             `json:"stackValue"`
	MaxStackValue      int             `json:"maxStackValue"`
	LastPlayedCard     Card            `json:"lastPlayedCard"`
}

func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		// decode json message
		var data map[string]interface{}
		err = json.Unmarshal(msg, &data)
		if err != nil {
			c.Conn.WriteJSON(Message{Error: "Invalid request body"})
			continue
		}

		action, exists := data["action"]
		if !exists {
			c.Conn.WriteJSON(Message{Error: "Invalid request body"})
			continue
		}

		switch action {
		case "join":
			isValid := true
			requiredFields := []string{"userId", "username", "profilePics"}
			for _, field := range requiredFields {
				if _, exists := data[field]; !exists {
					isValid = false
					break
				}
			}

			if !isValid {
				c.Conn.WriteJSON(Message{Error: "Invalid request body"})
				continue
			}

			c.ID = data["userId"].(string)

			// add player to the game
			newPlayer := &Player{
				Status:          "waiting",
				Cards:           []Card{},
				IsOut:           false,
				PlayerId:        c.ID,
				PlayerName:      data["username"].(string),
				PlayerAvatarURL: data["profilePics"].(string),
			}
			c.Pool.Game.Register <- newPlayer
			break
		case "start":
			fmt.Println("start")
			if c.Pool.Game.Status != "waiting" {
				c.Conn.WriteJSON(Message{Error: "Game has already started"})
				continue
			}

			if len(c.Pool.Game.Players) < 2 {
				c.Conn.WriteJSON(Message{Error: "Not enough players"})
				continue
			}

			c.Pool.Game.StartGame <- true

			break
		case "play":
			cardData, exists := data["card"]
			if !exists {
				c.Conn.WriteJSON(Message{Error: "Invalid request body"})
				continue
			}

			jsonData, _ := json.Marshal(cardData)

			var card Card
			json.Unmarshal(jsonData, &card)

			if !c.Pool.Game.isValidPlay(c.ID, card) {
				c.Conn.WriteJSON(Message{Error: "Invalid play"})
				continue
			}

			c.Pool.Game.cardPlayed <- card
			break
		case "leave":
			fmt.Println("leave")
			break
		default:
			c.Conn.WriteJSON(Message{Error: "Invalid action"})
		}

	}
}
