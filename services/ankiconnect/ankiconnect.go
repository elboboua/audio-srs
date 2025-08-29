package ankiconnect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AnkiConnectService struct {
	HttpClient http.Client
	Url        string
}

func (asc *AnkiConnectService) makeRequest(reqBody []byte) ([]byte, error) {
	req, err := http.NewRequest("GET", asc.Url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	resp, err := asc.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

type DecksResponse struct {
	Result []Deck
	Error  *string
}
type Deck string

func (acs *AnkiConnectService) GetDecks() ([]Deck, error) {
	reqBody := []byte(`{
	    "action": "deckNames",
	    "version": 6
	}`)

	resp, err := acs.makeRequest(reqBody)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewReader(resp))
	var decksResp DecksResponse
	err = decoder.Decode(&decksResp)
	if err != nil {
		return nil, err
	}

	return decksResp.Result, nil

}

type CardId int
type DueCardIdsResponse struct {
	Result []CardId
	Error  *string
}

func (acs *AnkiConnectService) GetDueCardIdsFromDeck(deck Deck) ([]CardId, error) {
	reqBody := fmt.Appendf([]byte{}, `{
		"action": "findCards",
		"version": 6,
		"params": {
		"query": "deck:%s (is:due OR is:new)"
		}
	}`, deck)

	resp, err := acs.makeRequest(reqBody)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewReader(resp))
	var dueCardIdsResp DueCardIdsResponse
	err = decoder.Decode(&dueCardIdsResp)
	if err != nil {
		return nil, err
	}

	return dueCardIdsResp.Result, nil
}

type FieldValue struct {
	Value string `json:"value"`
	Order int    `json:"order"`
}
type CardFields map[string]FieldValue

type Timestamp time.Time

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	var ts int64
	if err := json.Unmarshal(b, &ts); err != nil {
		return err
	}
	*t = Timestamp(time.Unix(ts, 0))
	return nil
}

type Card struct {
	CardId     CardId     `json:"cardId"`
	Fields     CardFields `json:"fields"`
	FieldOrder int        `json:"fieldOrder"`
	ModelName  string     `json:"modelName"`
	DeckName   Deck       `json:"deckName"`
	Due        Timestamp  `json:"due"`
}

type GetCardByIdResponse struct {
	Result []Card  `json:"result"`
	Error  *string `json:"error"`
}

func (acs *AnkiConnectService) GetCardById(cardId CardId) (Card, error) {
	reqBody := fmt.Appendf([]byte{}, `{
		 "action": "cardsInfo",
		 "version": 6,
		 "params": {
			"cards": [%d]
		}
	}`, cardId)

	resp, err := acs.makeRequest(reqBody)
	if err != nil {
		return Card{}, err
	}

	var cardByIdResp GetCardByIdResponse
	decoder := json.NewDecoder(bytes.NewReader(resp))
	err = decoder.Decode(&cardByIdResp)
	if err != nil {
		return Card{}, err
	}

	if len(cardByIdResp.Result) != 1 {
		return Card{}, fmt.Errorf("Unable to fetch card: id %d", cardId)
	}

	return cardByIdResp.Result[0], nil
}

// Ease is between 1 (Again) and 4 (Easy).
type Ease int

const (
	AgainEase Ease = 1
	HardEase  Ease = 2
	GoodEase  Ease = 3
	EasyEase  Ease = 4
)

type RepCardResponse struct {
	Result []bool  `json:"result"`
	Error  *string `json:"error"`
}

func (acs *AnkiConnectService) RepCard(cardId CardId, ease Ease) error {

	reqBody := fmt.Appendf([]byte{}, `{
		 "action": "answerCards",
		 "version": 6,
		 "params": {
			"answers": [
				{"cardId": %d, "ease": %d}
			]
		}
	}`, cardId, ease)

	resp, err := acs.makeRequest(reqBody)
	if err != nil {
		return err
	}

	var repCardResp RepCardResponse
	decoder := json.NewDecoder(bytes.NewBuffer(resp))
	err = decoder.Decode(&repCardResp)
	if err != nil {
		return err
	}

	if repCardResp.Error != nil {
		return fmt.Errorf("There was an error repping cardId %d: %s", cardId, *repCardResp.Error)
	} else if len(repCardResp.Result) > 0 && repCardResp.Result[0] == false {
		return fmt.Errorf("There was an unknown error repping cardId %d", cardId)
	}

	return nil
}
