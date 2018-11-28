package esi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/pequalsnp/go-eveonline/pkg/eveonline"
)

const CharacterWalletBalanceURLPattern = "https://esi.evetech.net/v1/characters/%d/wallet/"
const CharacterWalletJournalURLPattern = "https://esi.evetech.net/v4/characters/%d/wallet/journal/"

func (e *ESI) GetCharacterWalletBalance(authdClient *http.Client, characterID eveonline.CharacterID) (float64, error) {
	characterWalletURL := fmt.Sprintf(CharacterWalletBalanceURLPattern, characterID)

	resp, err := e.GetFromESI(characterWalletURL, authdClient, map[string][]string{})
	if err != nil {
		return 0.0, err
	}

	balance, err := strconv.ParseFloat(string(resp.Body), 64)
	if err != nil {
		return 0.0, fmt.Errorf("Failed to parse wallet balance for character id %d", characterID)
	}

	return balance, nil
}

type WalletTransaction struct {
}

func (e *ESI) GetCharacterWalletJournal(authdClient *http.Client, characterID eveonline.CharacterID) ([]*WalletTransaction, error) {
	characterWalletJournalURL := fmt.Sprintf(CharacterWalletJournalURLPattern, characterID)

	err := e.ScanPages(characterWalletJournalURL, authdClient, func(responsePage *ResponsePage) (bool, error) {
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
