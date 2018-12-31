package esi

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/pquerna/cachecontrol"
)

type ESICache interface {
	Put(key []byte, responsePage *ResponsePage) error
	Get(key []byte) (*ResponsePage, error)
}

type ESI struct {
	Cache      ESICache
	HttpClient *http.Client
}

func cacheKey(url string, queryParams map[string][]string) []byte {
	key := sha256.New()
	key.Write([]byte(url))
	var paramKeys []string
	for k := range queryParams {
		paramKeys = append(paramKeys, k)
	}
	sort.Strings(paramKeys)
	for _, paramKey := range paramKeys {
		key.Write([]byte(paramKey))
		paramValues := queryParams[paramKey]
		sort.Strings(paramValues)
		for _, paramValue := range paramValues {
			key.Write([]byte(paramValue))
		}
	}

	var keyBytes []byte
	return key.Sum(keyBytes)
}

func (e *ESI) CacheResponsePage(url string, queryParams map[string][]string, responsePage *ResponsePage) error {
	key := cacheKey(url, queryParams)
	return e.Cache.Put(key, responsePage)
}

func (e *ESI) GetFromCache(url string, queryParams map[string][]string) (*ResponsePage, error) {
	key := cacheKey(url, queryParams)
	return e.Cache.Get(key)
}

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

func (e *ESI) GetFromESI(url string, httpClient *http.Client, queryParams map[string][]string) (*ResponsePage, error) {
	if httpClient == nil {
		httpClient = e.HttpClient
	}
	cachedPage, err := e.GetFromCache(url, queryParams)
	var previousEtag *string
	if err == nil && cachedPage != nil {
		if !cachedPage.Expired() {
			fmt.Printf("%s %v cache hit\n", url, queryParams)
			return cachedPage, nil
		}
		previousEtag = &cachedPage.Etag
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if previousEtag != nil {
		request.Header.Add("If-None-Match", *previousEtag)
	}

	query := request.URL.Query()
	for param, vals := range queryParams {
		for _, val := range vals {
			query.Add(param, val)
		}
	}
	request.URL.RawQuery = query.Encode()
	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body []byte
	if resp.StatusCode == 304 {
		fmt.Printf("%s %v etag match response\n", url, queryParams)
		body = cachedPage.Body
	} else {
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
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

func (e *ESI) ScanPages(url string, httpClient *http.Client, scanFn func(*ResponsePage) (bool, error)) error {
	page := 1
	for {
		responsePage, err := e.GetFromESI(url, httpClient, map[string][]string{"page": []string{strconv.Itoa(page)}})
		if err != nil {
			return err
		}

		continueScan, err := scanFn(responsePage)
		if err != nil {
			return err
		}
		if !continueScan {
			break
		}

		pagesStr := responsePage.Headers.Get("x-pages")
		pages := 0
		if pagesStr != "" {
			pagesConverted, err := strconv.Atoi(pagesStr)
			if err != nil {
				return err
			}
			pages = pagesConverted
		}

		page = page + 1
		if page >= pages {
			break
		}
	}

	return nil
}

func (e *ESI) GetAllPages(
	url string,
	page int,
	queryParams map[string][]string,
	httpClient *http.Client,
) ([]*ResponsePage, error) {
	params := map[string][]string{"page": []string{strconv.Itoa(page)}}
	for k, v := range queryParams {
		params[k] = v
	}
	responsePage, err := e.GetFromESI(url, httpClient, params)
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
		nextPages, nextPageErr := e.GetAllPages(url, page+1, queryParams, httpClient)
		if nextPageErr != nil {
			return nil, nextPageErr
		}

		return append(responsePages, nextPages...), nil

	}

	return responsePages, nil
}
