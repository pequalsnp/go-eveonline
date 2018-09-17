package eveonline

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/pquerna/cachecontrol"
)

type CategoryID int64
type GroupID int64
type RegionID int64
type TypeID int64
type StationID int
type SystemID int
type ConstellationID int
type LocationID int64

type CacheInfo struct {
	ExpiresAt time.Time
	Etag      string
}

type ResponsePage struct {
	CacheInfo
	Body               []byte
	ResponseStatusCode int
	Headers            http.Header
}

func (c CacheInfo) Expired() bool {
	return c.ExpiresAt.Before(time.Now())
}

func GetFromESI(url string, httpClient *http.Client, queryParams map[string][]string) (*ResponsePage, error) {
	request, newRequestErr := http.NewRequest("GET", url, nil)
	if newRequestErr != nil {
		return nil, newRequestErr
	}
	query := request.URL.Query()
	for param, vals := range queryParams {
		for _, val := range vals {
			query.Add(param, val)
		}
	}
	request.URL.RawQuery = query.Encode()
	resp, requestErr := httpClient.Do(request)
	if requestErr != nil {
		return nil, requestErr
	}
	defer resp.Body.Close()

	body, bodyReadErr := ioutil.ReadAll(resp.Body)
	if bodyReadErr != nil {
		return nil, bodyReadErr
	}

	etag := resp.Header.Get("etag")

	_, expires, cachecontrolParseError := cachecontrol.CachableResponse(request, resp, cachecontrol.Options{})
	if cachecontrolParseError != nil {
		return nil, cachecontrolParseError
	}

	return &ResponsePage{
		CacheInfo:          CacheInfo{Etag: etag, ExpiresAt: expires},
		Body:               body,
		ResponseStatusCode: resp.StatusCode,
		Headers:            resp.Header,
	}, nil
}

func GetAllPages(url string, page int, httpClient *http.Client) ([]*ResponsePage, error) {
	responsePage, err := GetFromESI(url, httpClient, map[string][]string{"page": []string{strconv.Itoa(page)}})
	if err != nil {
		return nil, err
	}

	pagesStr := responsePage.Headers.Get("x-pages")
	pages := 0
	if pagesStr != "" {
		pagesConverted, pagesStringErr := strconv.Atoi(pagesStr)
		if pagesStringErr != nil {
			return nil, pagesStringErr
		}
		pages = pagesConverted
	}

	responsePages := []*ResponsePage{responsePage}

	if page < pages {
		nextPages, nextPageErr := GetAllPages(url, page+1, httpClient)
		if nextPageErr != nil {
			return nil, nextPageErr
		}

		return append(responsePages, nextPages...), nil

	}

	return responsePages, nil
}
