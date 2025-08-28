package main

import (
	"fmt"
	"net/http"

	"github.com/elboboua/audio-srs/services/ankiconnect"
)

func main() {
	fmt.Println("Hello, world!")
	acs := ankiconnect.AnkiConnectService{
		HttpClient: *http.DefaultClient,
		Url:        "http://localhost:8765",
	}

	decks, err := acs.GetDecks()
	if err != nil {
		panic(err)
	}

	for i, deck := range decks {
		fmt.Printf("%2d: %s\n", i, deck)
	}

	cardIds, err := acs.GetDueCardIdsFromDeck(decks[5])
	if err != nil {
		panic(err)
	}
	for i, cardId := range cardIds {
		fmt.Printf("%2d: %d\n", i, cardId)
	}

	card, err := acs.GetCardById(cardIds[0])
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", card)
}
