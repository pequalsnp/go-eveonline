package esi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pequalsnp/go-eveonline/pkg/eveonline"
)

type Order struct {
	ID         int64            `json:"order_id"`
	IsBuy      bool             `json:"is_buy_order"`
	LocationID int64            `json:"location_id"`
	Price      float64          `json:"price"`
	TypeID     eveonline.TypeID `json:"type_id"`
}

type Orders struct {
	RegionID  eveonline.RegionID
	Orders    []*Order
	ExpiresAt time.Time
}

func GetOrders(regionID eveonline.RegionID, httpClient *http.Client) (*Orders, error) {
	regionMarketOrdersURL := fmt.Sprintf("https://esi.evetech.net/v1/markets/%d/orders/", regionID)

	allPages, err := eveonline.GetAllPages(regionMarketOrdersURL, 1, httpClient)
	if err != nil {
		return nil, err
	}

	var latestExpiry time.Time
	orders := make([]*Order, 0)
	for _, rp := range allPages {
		if rp.ExpiresAt.After(latestExpiry) {
			latestExpiry = rp.ExpiresAt
		}

		unmarshalErr := json.Unmarshal(rp.Body, &orders)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}
	}

	return &Orders{RegionID: regionID, Orders: orders, ExpiresAt: latestExpiry}, nil
}
