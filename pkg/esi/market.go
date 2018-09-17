package esi

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/pequalsnp/go-eveonline/pkg/eveonline"
)

type Order struct {
	ID         int64                `json:"order_id"`
	IsBuy      bool                 `json:"is_buy_order"`
	LocationID eveonline.LocationID `json:"location_id"`
	Price      float64              `json:"price"`
	TypeID     eveonline.TypeID     `json:"type_id"`
}

type Orders struct {
	RegionID    eveonline.RegionID
	Orders      []*Order
	ExpiresAt   time.Time
	HighestBuys map[eveonline.TypeID]float64
	LowestSells map[eveonline.TypeID]float64
}

func GetOrders(regionID eveonline.RegionID, locationID *eveonline.LocationID, httpClient *http.Client) (*Orders, error) {
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

	if locationID != nil {
		unfilteredOrders := orders
		orders = make([]*Order, 0, len(unfilteredOrders))
		for _, order := range unfilteredOrders {
			if order.LocationID == *locationID {
				orders = append(orders, order)
			}
		}
	}

	lowestSell := make(map[eveonline.TypeID]float64, 0)
	highestBuy := make(map[eveonline.TypeID]float64, 0)
	for _, order := range orders {
		if order.IsBuy {
			previousHighestBuy := 0.0
			phb, ok := highestBuy[order.TypeID]
			if ok {
				previousHighestBuy = phb
			}
			highestBuy[order.TypeID] = math.Max(order.Price, previousHighestBuy)
		} else {
			previousLowestSell := math.MaxFloat64
			pls, ok := lowestSell[order.TypeID]
			if ok {
				previousLowestSell = pls
			}
			lowestSell[order.TypeID] = math.Min(order.Price, previousLowestSell)
		}
	}

	return &Orders{RegionID: regionID, Orders: orders, ExpiresAt: latestExpiry, HighestBuys: highestBuy, LowestSells: lowestSell}, nil
}
