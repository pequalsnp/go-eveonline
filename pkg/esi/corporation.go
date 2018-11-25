package esi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pequalsnp/go-eveonline/pkg/eveonline"
)

type Corporation struct {
	Name        string                  `json:"name"`
	ID          eveonline.CorporationID `json:"corporation_id"`
	Ticker      string                  `json:"ticker"`
	MemberCount int                     `json:"member_count"`
}

const CorporationInfoURLPattern = "/v4/corporations/%d/"

func GetCorporation(corporationID eveonline.CorporationID, httpClient *http.Client) (*Corporation, error) {
	url := fmt.Sprintf(CorporationInfoURLPattern, corporationID)

	resp, err := eveonline.GetFromESI(url, httpClient, nil)
	if err != nil {
		return nil, err
	}

	corporation := new(Corporation)
	err = json.Unmarshal(resp.Body, corporation)
	if err != nil {
		return nil, err
	}

	return corporation, nil
}
