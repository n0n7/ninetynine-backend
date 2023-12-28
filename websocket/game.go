package websocket

import (
	"fmt"
	"math/rand"
	Room "ninetynine/room"
	"time"
)

type Game struct {
	Players            []*Player
	Status             string
	CurrentPlayerIndex int
	CurrentDirection   int
	StackValue         int
	MaxStackValue      int
	CardPerPlayer      int
	LastPlayedCard     Card
	Register           chan *Player
	Unregister         chan string
	cardPlayed         chan Card
	StartGame          chan bool
	Stop               chan bool
	Pool               *Pool
}

type Player struct {
	Status          string
	Cards           []Card
	IsOut           bool
	PlayerId        string
	PlayerName      string
	PlayerAvatarURL string
}

type Card struct {
	Value     int  `json:"value"`
	IsSpecial bool `json:"isSpecial"`
}

func NewGame() *Game {
	return &Game{
		Players:            []*Player{},
		Status:             "waiting",
		CurrentPlayerIndex: 0,
		CurrentDirection:   1,
		StackValue:         0,
		MaxStackValue:      99,
		CardPerPlayer:      3,
		LastPlayedCard: Card{
			Value:     -1,
			IsSpecial: true,
		}, // empty card
		Register:   make(chan *Player),
		Unregister: make(chan string),
		cardPlayed: make(chan Card),
		StartGame:  make(chan bool),
		Stop:       make(chan bool),
	}
}

func (game *Game) Start() {
	defer func() {
		fmt.Println("game stopped")
		game.Pool.GameAction <- "game ended"
		Room.FirebaseUpdateChannel[game.Pool.RoomId] <- Room.FirebaseUpdateData{
			Field: "status",
			Value: game.Status,
		}
	}()

	for {
		select {
		case stop := <-game.Stop:
			if stop {
				return
			}

		case player := <-game.Register:
			game.Players = append(game.Players, player)
			fmt.Println("register player", player.PlayerId)
			game.Pool.GameAction <- fmt.Sprintf("player %v joined", player.PlayerName)

		case playerId := <-game.Unregister:
			if game.Status == "playing" || game.Status == "ended" {
				for _, p := range game.Players {
					if p.PlayerId == playerId {
						p.IsOut = true
						p.Status = "left"
						fmt.Println("unregister player from the game", p.PlayerId)
						game.Pool.GameAction <- fmt.Sprintf("player %v left", p.PlayerName)
						break
					}
				}

				if playerId == game.Players[game.CurrentPlayerIndex].PlayerId {
					game.NextPlayer()
				} else {
					if game.IsGameEnded() {
						game.Status = "ended"
						fmt.Println("game ended")
						go func() {
							game.Stop <- true
						}()
						return
					}
				}
				break
			}

			if game.Status == "waiting" {
				for i, p := range game.Players {
					if p.PlayerId == playerId {
						game.Players = append(game.Players[:i], game.Players[i+1:]...)
						fmt.Println("unregister player from the game", p.PlayerId)
						game.Pool.GameAction <- fmt.Sprintf("player %v left", p.PlayerName)
						break
					}
				}
			}
			break

		case _ = <-game.StartGame:
			game.Status = "playing"
			for _, p := range game.Players {
				p.Cards = []Card{}
				p.Status = "playing"
				for i := 0; i < game.CardPerPlayer; i++ {
					p.Cards = append(p.Cards, randomCard())
				}
			}

			game.Pool.GameAction <- "game started"
			Room.FirebaseUpdateChannel[game.Pool.RoomId] <- Room.FirebaseUpdateData{
				Field: "status",
				Value: game.Status,
			}
			if !game.CanCurrentPlayerPlay() {
				game.NextPlayer()
			}
		case card := <-game.cardPlayed:
			fmt.Println("card played", card)
			game.LastPlayedCard = card
			player := game.Players[game.CurrentPlayerIndex]
			game.PlayCard(card)
			game.NextPlayer()

			if !game.IsGameEnded() {
				game.Pool.GameAction <- fmt.Sprintf("player %v played Card%v", player.PlayerName, card)
			}
		}

	}
}

func (game *Game) PlayCard(card Card) {
	if !card.IsSpecial {
		game.StackValue += card.Value
	} else {
		switch card.Value {
		case 0:
			break
		case 1:
			game.CurrentDirection *= -1
			break
		case 2:
			game.shufflePlayer()
			break
		case 3:
			game.StackValue = game.MaxStackValue
			break
		}
	}

	// find card index and change it to new card
	for i, c := range game.Players[game.CurrentPlayerIndex].Cards {
		if c.Value == card.Value && c.IsSpecial == card.IsSpecial {
			game.Players[game.CurrentPlayerIndex].Cards[i] = randomCard()
			break
		}
	}
}

func (game *Game) NextPlayer() {
	game.CurrentPlayerIndex += game.CurrentDirection
	if game.CurrentPlayerIndex < 0 {
		game.CurrentPlayerIndex = len(game.Players) - 1
	} else if game.CurrentPlayerIndex >= len(game.Players) {
		game.CurrentPlayerIndex = 0
	}

	for range game.Players {
		if game.IsGameEnded() {
			fmt.Println("game ended")
			game.Status = "ended"
			go func() {
				game.Stop <- true
			}()
			return
		}

		if game.CanCurrentPlayerPlay() {
			break
		}

		if !game.Players[game.CurrentPlayerIndex].IsOut {
			fmt.Println("player", game.Players[game.CurrentPlayerIndex].PlayerName, "is out")
			game.Players[game.CurrentPlayerIndex].IsOut = true
			game.Players[game.CurrentPlayerIndex].Status = "Out"
			game.Pool.GameAction <- fmt.Sprintf("player %v is out", game.Players[game.CurrentPlayerIndex].PlayerName)
		}

		game.CurrentPlayerIndex += game.CurrentDirection
		if game.CurrentPlayerIndex < 0 {
			game.CurrentPlayerIndex = len(game.Players) - 1
		} else if game.CurrentPlayerIndex >= len(game.Players) {
			game.CurrentPlayerIndex = 0
		}

	}
}

func (game *Game) IsGameEnded() bool {
	return game.currentPlayerCount() == 1
}

func (game *Game) currentPlayerCount() int {
	count := 0
	for _, p := range game.Players {
		if !p.IsOut {
			count++
		}
	}
	return count
}

func (game *Game) CanCurrentPlayerPlay() bool {
	if game.Players[game.CurrentPlayerIndex].IsOut {
		return false
	}

	for _, card := range game.Players[game.CurrentPlayerIndex].Cards {
		if game.isValidPlay(game.Players[game.CurrentPlayerIndex].PlayerId, card) {
			return true
		}
	}
	return false
}

func (game *Game) shufflePlayer() {
	currentPlayerId := game.Players[game.CurrentPlayerIndex].PlayerId
	players := game.Players
	rand.NewSource(time.Now().UnixNano())
	rand.Shuffle(len(players), func(i, j int) { players[i], players[j] = players[j], players[i] })

	for i, p := range players {
		if p.PlayerId == currentPlayerId {
			game.CurrentPlayerIndex = i
			break
		}
	}
}

func (game *Game) GetGameData(userId string) GameMessage {
	gameData := GameMessage{
		Players:            []PlayerMessage{},
		PlayerCards:        []Card{},
		Status:             game.Status,
		CurrentPlayerIndex: game.CurrentPlayerIndex,
		CurrentDirection:   game.CurrentDirection,
		StackValue:         game.StackValue,
		MaxStackValue:      game.MaxStackValue,
		LastPlayedCard:     game.LastPlayedCard,
	}

	for _, p := range game.Players {
		gameData.Players = append(gameData.Players, getPlayerData(*p))
		if p.PlayerId == userId {
			gameData.PlayerCards = p.Cards
		}
	}

	return gameData
}

func getPlayerData(p Player) PlayerMessage {
	return PlayerMessage{
		PlayerId:        p.PlayerId,
		PlayerName:      p.PlayerName,
		PlayerAvatarURL: p.PlayerAvatarURL,
		IsOut:           p.IsOut,
		Status:          p.Status,
	}
}

func (game *Game) isValidPlay(playerId string, card Card) bool {
	// check if it's player turn
	if game.Players[game.CurrentPlayerIndex].PlayerId != playerId {
		fmt.Println("not player turn")
		return false
	}
	// check if player has that card
	hasCard := false
	for _, c := range game.Players[game.CurrentPlayerIndex].Cards {
		if c.Value == card.Value && c.IsSpecial == card.IsSpecial {
			hasCard = true
			break
		}
	}

	if !hasCard {
		fmt.Println("player doesn't have that card")
		return false
	}

	// check if card is valid
	if card.IsSpecial {
		return true
	}

	// check if card value is valid
	if card.Value+game.StackValue <= game.MaxStackValue {
		return true
	}

	fmt.Println("card value is invalid")
	return false
}

func randomCard() Card {
	cardListLength := len(CardList)
	rand.NewSource(time.Now().UnixNano())
	index := rand.Intn(cardListLength)
	return CardList[index]
}

var CardList = [16]Card{
	{Value: 1, IsSpecial: false},
	{Value: 2, IsSpecial: false},
	{Value: 3, IsSpecial: false},
	{Value: 4, IsSpecial: false},
	{Value: 5, IsSpecial: false},
	{Value: 6, IsSpecial: false},
	{Value: 7, IsSpecial: false},
	{Value: 8, IsSpecial: false},
	{Value: 9, IsSpecial: false},
	{Value: 10, IsSpecial: false},
	{Value: -9, IsSpecial: false},
	{Value: -10, IsSpecial: false},
	{Value: 0, IsSpecial: true},
	{Value: 1, IsSpecial: true},
	{Value: 2, IsSpecial: true},
	{Value: 3, IsSpecial: true},
}
