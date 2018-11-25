package esi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pequalsnp/go-eveonline/pkg/eveonline"
)

type CorporationLoyaltyPoints map[eveonline.CorporationID]int64

type esiLoaytyPointEntry struct {
	CorporationID eveonline.CorporationID
	LocaltyPoints int64
}

const CharacterLoyaltyPointsURLPattern = "/v1/characters/%d/loyalty/points/"

func GetLoyaltyPoints(characterID *eveonline.CharacterID, authdClient *http.Client) (CorporationLoyaltyPoints, error) {
	url := fmt.Sprintf(CharacterLoyaltyPointsURLPattern, characterID)

	resp, err := eveonline.GetFromESI(url, authdClient, nil)
	if err != nil {
		return nil, err
	}

	esiLoyaltyPointEntries := make([]esiLoaytyPointEntry, 0)
	err = json.Unmarshal(resp.Body, esiLoyaltyPointEntries)
	if err != nil {
		return nil, err
	}

	corporationLoyaltyPoints := make(CorporationLoyaltyPoints)
	for _, esiEntry := range esiLoyaltyPointEntries {
		corporationLoyaltyPoints[esiEntry.CorporationID] = esiEntry.LocaltyPoints
	}

	return corporationLoyaltyPoints, nil
}
