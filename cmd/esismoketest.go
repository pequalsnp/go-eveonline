package main

import (
	"log"
	"net/http"

	"github.com/pequalsnp/go-eveonline/pkg/esi"
)

type SearchSmoketest struct {
	Query      string
	Categories []esi.SearchCategory
}

func main() {
	httpClient := &http.Client{}
	searchSmoketests := []SearchSmoketest{
		SearchSmoketest{
			Query:      "Capital Ships",
			Categories: []esi.SearchCategory{esi.InventoryTypeSearchCategory{}}},
	}
	for _, searchSmoketest := range searchSmoketests {
		log.Printf(
			"Performing search smoketest query: '%s' categories %v\n",
			searchSmoketest.Query,
			searchSmoketest.Categories,
		)
		result, err := esi.Search(searchSmoketest.Query, searchSmoketest.Categories, true, httpClient)
		if err != nil {
			log.Printf("failed search smoketest query: '%s' categories %v error: %v\n",
				searchSmoketest.Query,
				searchSmoketest.Categories,
				err,
			)
			panic(err)
		}
		log.Printf("Result: %v", result)
		for _, category := range searchSmoketest.Categories {
			for _, resultObj := range result[category.ApiName()] {
				switch resultObj.(type) {
				case *esi.Type:
					typeObj := resultObj.(*esi.Type)
					log.Println(typeObj)
				}
			}
		}

	}

}
