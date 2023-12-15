package websocket

import (
	"fmt"
	"math/rand"
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
	Unregister         chan *Player
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
		Unregister: make(chan *Player),
		cardPlayed: make(chan Card),
		StartGame:  make(chan bool),
		Stop:       make(chan bool),
	}
}

func (game *Game) Start() {
	defer func() {
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

		case _ = <-game.StartGame:
			game.Status = "playing"
			for _, p := range game.Players {
				for i := 0; i < game.CardPerPlayer; i++ {
					p.Cards = append(p.Cards, randomCard())
				}
			}
			game.Pool.GameAction <- "game started"
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

	fmt.Println("game data", gameData)

	return gameData
}

func getPlayerData(p Player) PlayerMessage {
	return PlayerMessage{
		PlayerId:        p.PlayerId,
		PlayerName:      p.PlayerName,
		PlayerAvatarURL: p.PlayerAvatarURL,
	}
}

func (game *Game) isValidPlay(playerId string, card Card) bool {
	// check if it's player turn
	if game.Players[game.CurrentPlayerIndex].PlayerId != playerId {
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

	return false
}

func randomCard() Card {
	cardListLength := len(CardList)
	rand.NewSource(time.Now().UnixNano())
	index := rand.Intn(cardListLength)
	return CardList[index]
}

var CardList = [14]Card{
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
	{Value: 0, IsSpecial: true},
	{Value: 1, IsSpecial: true},
	{Value: 2, IsSpecial: true},
	{Value: 3, IsSpecial: true},
}
