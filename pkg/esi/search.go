package esi

import (
	"encoding/json"

	"github.com/pequalsnp/go-eveonline/pkg/eveonline"
)

type SearchCategory interface {
	ApiName() string
}

type InventoryTypeSearchCategory struct {
	FilterCategoryID *eveonline.CategoryID
}

func (_ InventoryTypeSearchCategory) ApiName() string {
	return "inventory_type"
}

const SearchURL = "https://esi.evetech.net/latest/search/"

type SearchResults map[string][]interface{}

func (e *ESI) Search(query string, categories []SearchCategory, strict bool) (SearchResults, error) {
	categoryNames := make([]string, 0, len(categories))
	for _, category := range categories {
		categoryNames = append(categoryNames, category.ApiName())
	}

	params := map[string][]string{"search": []string{query}, "categories": categoryNames}
	if strict {
		params["strict"] = []string{"true"}
	}

	resp, err := e.GetFromESI(SearchURL, nil, params)
	if err != nil {
		return nil, err
	}

	results := make(map[string][]int64)
	err = json.Unmarshal(resp.Body, &results)
	if err != nil {
		return nil, err
	}

	searchResults := make(SearchResults)
	for _, category := range categories {
		resultsForCategory, ok := results[category.ApiName()]
		if !ok {
			continue
		}

		switch category.(type) {
		case InventoryTypeSearchCategory:
			for _, typeID := range resultsForCategory {
				if err != nil {
					return nil, err
				}
				typeObj, err := e.GetType(eveonline.TypeID(typeID))
				if err != nil {
					return nil, err
				}

				if !typeObj.Published {
					continue
				}

				optionalFilterCategoryID := category.(InventoryTypeSearchCategory).FilterCategoryID
				if optionalFilterCategoryID != nil {
					groupObj, err := e.GetGroup(typeObj.GroupID)
					if err != nil {
						return nil, err
					}

					if groupObj.CategoryID != *optionalFilterCategoryID {
						continue
					}
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
