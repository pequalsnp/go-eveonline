package esi

import (
	"encoding/json"
	"net/http"

	"github.com/pequalsnp/go-eveonline/pkg/eveonline"
)

type SearchCategory interface {
	ApiName() string
}

type InventoryTypeSearchCategory struct{}

func (c InventoryTypeSearchCategory) ApiName() string {
	return "inventory_type"
}

const SearchURL = "https://esi.evetech.net/v2/search/"

type SearchResults map[string][]interface{}

func Search(query string, categories []SearchCategory, httpClient *http.Client) (SearchResults, error) {
	categoryNames := make([]string, 0, len(categories))
	for _, category := range categories {
		categoryNames = append(categoryNames, category.ApiName())
	}

	resp, err := eveonline.GetFromESI(SearchURL, httpClient, map[string][]string{"categories": categoryNames})
	if err != nil {
		return nil, err
	}

	results := make(map[string][]int64)
	err = json.Unmarshal(resp.Body, &results)
	if err != nil {
		return nil, err
	}

	universe := GetUniverse()
	searchResults := make(SearchResults)
	for _, category := range categories {
		resultsForCategory, ok := results[category.ApiName()]
		if !ok {
			continue
		}

		switch category.(type) {
		case InventoryTypeSearchCategory:
			for _, typeID := range resultsForCategory {
				typeObj, err := universe.GetType(eveonline.TypeID(typeID), httpClient)
				if err != nil {
					return nil, err
				}
				categorySearchResults, ok := searchResults[category.ApiName()]
				if !ok {
					searchResults[category.ApiName()] = make([]interface{}, 0)
				}
				searchResults[category.ApiName()] = append(categorySearchResults, typeObj)
			}
		}
	}

	return searchResults, nil
}
