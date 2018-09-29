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
	RegionID  eveonline.RegionID
	Orders    []*Order
	ExpiresAt time.Time
}

type Market struct {
	RegionID    eveonline.RegionID
	ExpiresAt   time.Time
	HighestBuys map[eveonline.TypeID]float64
	LowestSells map[eveonline.TypeID]float64
}

type esiMarketPrice struct {
	TypeID        eveonline.TypeID `json:"type_id"`
	AveragePrice  float64          `json:"average_price"`
	AdjustedPrice float64          `json:"adjusted_price"`
}

type AveragePrices map[eveonline.TypeID]float64

func GetMarket(regionID eveonline.RegionID, locationID *eveonline.LocationID, httpClient *http.Client) (*Market, error) {
	regionMarketOrdersURL := fmt.Sprintf("https://esi.evetech.net/v1/markets/%d/orders/", regionID)
	var latestExpiry time.Time
	market := &Market{
		RegionID:    regionID,
		HighestBuys: make(map[eveonline.TypeID]float64),
		LowestSells: make(map[eveonline.TypeID]float64),
	}
	err := eveonline.ScanPages(regionMarketOrdersURL, httpClient, func(page *eveonline.ResponsePage) (bool, error) {
		if page.ExpiresAt.After(latestExpiry) {
			latestExpiry = page.ExpiresAt
		}

		orders := make([]*Order, 0)
		unmarshalErr := json.Unmarshal(page.Body, &orders)
		if unmarshalErr != nil {
			return false, unmarshalErr
		}

		for _, order := range orders {
			if order.IsBuy {
				previousHighestBuy := 0.0
				phb, ok := market.HighestBuys[order.TypeID]
				if ok {
					previousHighestBuy = phb
				}
				market.HighestBuys[order.TypeID] = math.Max(order.Price, previousHighestBuy)
			} else {
				previousLowestSell := math.MaxFloat64
				pls, ok := market.LowestSells[order.TypeID]
				if ok {
					previousLowestSell = pls
				}
				market.LowestSells[order.TypeID] = math.Min(order.Price, previousLowestSell)
			}
		}

		return true, nil
	})
	if err != nil {
		return nil, err
	}

	market.ExpiresAt = latestExpiry
	return market, nil
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

	return &Orders{RegionID: regionID, Orders: orders, ExpiresAt: latestExpiry}, nil
}

func GetAverageMarketPrices(httpClient *http.Client) (AveragePrices, error) {
	marketPricesURL := "https://esi.evetech.net/v1/markets/prices/"
	resp, err := eveonline.GetFromESI(marketPricesURL, httpClient, map[string][]string{})
	if err != nil {
		return nil, err
	}

	averagePriceList := make([]*esiMarketPrice, 0)
	err = json.Unmarshal(resp.Body, &averagePriceList)
	if err != nil {
		return nil, err
	}

	averagePrices := make(AveragePrices)
	for _, esiPrice := range averagePriceList {
		averagePrices[esiPrice.TypeID] = esiPrice.AveragePrice
	}

	return averagePrices, nil
}
