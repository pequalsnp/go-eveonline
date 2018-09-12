package esi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/pequalsnp/go-eveonline/pkg/eveonline"
)

type Type struct {
	eveonline.CacheInfo
	ID        eveonline.TypeID  `json:"type_id"`
	GroupID   eveonline.GroupID `json:"group_id"`
	Volume    float64           `json:"volume"`
	Name      string            `json:"name"`
	Published bool              `json:"published"`
}

type Group struct {
	eveonline.CacheInfo
	ID         eveonline.GroupID    `json:"group_id"`
	CategoryID eveonline.CategoryID `json:"category_id"`
	Types      []eveonline.TypeID   `json:"types"`
	Name       string               `json:"name"`
	Published  bool                 `json:"published"`
}

type Category struct {
	eveonline.CacheInfo
	ID        eveonline.CategoryID `json:"category_id"`
	Groups    []eveonline.GroupID  `json:"groups"`
	Name      string               `json:"name"`
	Published bool                 `json:"published"`
}

type Universe struct {
	lock  *sync.RWMutex
	types map[eveonline.TypeID]*Type
}

var currentUniverse *Universe
var once sync.Once

func GetUniverse() *Universe {
	once.Do(func() {
		currentUniverse = &Universe{lock: new(sync.RWMutex), types: make(map[eveonline.TypeID]*Type)}
	})
	return currentUniverse
}

const TypeURLPattern = "https://esi.evetech.net/v3/universe/types/%d/"

func (u *Universe) GetType(typeID eveonline.TypeID, httpClient *http.Client) (*Type, error) {
	u.lock.RLock()
	typeObj, ok := u.types[typeID]
	if ok && !typeObj.Expired() {
		u.lock.RUnlock()
		return typeObj, nil
	}
	resp, err := eveonline.GetFromESI(
		fmt.Sprintf(TypeURLPattern, typeID),
		httpClient,
		map[string][]string{},
	)
	if err != nil {
		u.lock.RUnlock()
		return nil, err
	}

	typeObj = new(Type)
	err = json.Unmarshal(resp.Body, &typeObj)
	if err != nil {
		u.lock.RUnlock()
		return nil, err
	}
	typeObj.ExpiresAt = resp.ExpiresAt
	typeObj.Etag = resp.Etag

	u.lock.RUnlock()
	u.lock.Lock()
	defer u.lock.Unlock()
	potentiallyRefreshedTypeObj, ok := u.types[typeID]
	if ok && !potentiallyRefreshedTypeObj.Expired() {
		// Refreshed while upgrading the lock, so return the new version
		return potentiallyRefreshedTypeObj, nil
	}

	u.types[typeID] = typeObj

	return typeObj, nil
}
