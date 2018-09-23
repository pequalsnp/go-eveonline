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

type Station struct {
	eveonline.CacheInfo
	ID              eveonline.StationID       `json:"station_id"`
	Name            string                    `json:"name"`
	SystemID        eveonline.SystemID        `json:"system_id"`
	ConstellationID eveonline.ConstellationID `json:",omitempty"`
	RegionID        eveonline.RegionID        `json:",omitempty"`
}

type System struct {
	eveonline.CacheInfo
	ID              eveonline.SystemID        `json:"system_id"`
	ConstellationID eveonline.ConstellationID `json:"constellation_id"`
}

type Constellation struct {
	eveonline.CacheInfo
	ID       eveonline.ConstellationID `json:"constellation_id"`
	RegionID eveonline.RegionID        `json:"region_id"`
}

type Universe struct {
	lock           *sync.RWMutex
	types          map[eveonline.TypeID]*Type
	stations       map[eveonline.StationID]*Station
	systems        map[eveonline.SystemID]*System
	constellations map[eveonline.ConstellationID]*Constellation
}

var currentUniverse *Universe
var once sync.Once

func GetUniverse() *Universe {
	once.Do(func() {
		currentUniverse = &Universe{
			lock:           new(sync.RWMutex),
			types:          make(map[eveonline.TypeID]*Type),
			stations:       make(map[eveonline.StationID]*Station),
			systems:        make(map[eveonline.SystemID]*System),
			constellations: make(map[eveonline.ConstellationID]*Constellation),
		}
	})
	return currentUniverse
}

const TypeURLPattern = "https://esi.evetech.net/v3/universe/types/%d/"
const StationURLPattern = "https://esi.evetech.net/v2/universe/stations/%d/"
const SystemURLPattern = "https://esi.evetech.net/v4/universe/systems/%d/"
const ConstellationURLPattern = "https://esi.evetech.net/v1/universe/constellations/%d/"

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

func (u *Universe) GetStation(stationID eveonline.StationID, httpClient *http.Client) (*Station, error) {
	u.lock.RLock()
	stationObj, ok := u.stations[stationID]
	if ok && !stationObj.Expired() {
		u.lock.RUnlock()
		return stationObj, nil
	}

	resp, err := eveonline.GetFromESI(
		fmt.Sprintf(StationURLPattern, stationID),
		httpClient,
		map[string][]string{},
	)
	if err != nil {
		u.lock.RUnlock()
		return nil, err
	}

	stationObj = new(Station)
	err = json.Unmarshal(resp.Body, &stationObj)
	if err != nil {
		u.lock.RUnlock()
		return nil, err
	}
	stationObj.ExpiresAt = resp.ExpiresAt
	stationObj.Etag = resp.Etag

	u.lock.RUnlock()

	system, err := u.GetSystem(stationObj.SystemID, httpClient)
	if err != nil {
		return nil, err
	}

	constellation, err := u.GetConstellation(system.ConstellationID, httpClient)
	if err != nil {
		return nil, err
	}

	u.lock.Lock()
	defer u.lock.Unlock()
	potentiallyRefreshedStationObj, ok := u.stations[stationID]
	if ok && !potentiallyRefreshedStationObj.Expired() {
		// Refreshed while upgrading the lock, so return the new version
		return potentiallyRefreshedStationObj, nil
	}

	stationObj.ConstellationID = constellation.ID
	stationObj.RegionID = constellation.RegionID

	u.stations[stationID] = stationObj

	return stationObj, nil
}

func (u *Universe) GetSystem(systemID eveonline.SystemID, httpClient *http.Client) (*System, error) {
	u.lock.RLock()
	systemObj, ok := u.systems[systemID]
	if ok && !systemObj.Expired() {
		u.lock.RUnlock()
		return systemObj, nil
	}

	resp, err := eveonline.GetFromESI(
		fmt.Sprintf(SystemURLPattern, systemID),
		httpClient,
		map[string][]string{},
	)
	if err != nil {
		u.lock.RUnlock()
		return nil, err
	}

	systemObj = new(System)
	err = json.Unmarshal(resp.Body, systemObj)
	if err != nil {
		u.lock.RUnlock()
		return nil, err
	}
	systemObj.ExpiresAt = resp.ExpiresAt
	systemObj.Etag = resp.Etag

	u.lock.RUnlock()
	u.lock.Lock()
	defer u.lock.Unlock()
	potentiallyRefreshedSystemObj, ok := u.systems[systemID]
	if ok && !potentiallyRefreshedSystemObj.Expired() {
		// Refreshed while upgrading the lock, so return the new version
		return potentiallyRefreshedSystemObj, nil
	}

	u.systems[systemID] = systemObj

	return systemObj, nil
}

func (u *Universe) GetConstellation(constellationID eveonline.ConstellationID, httpClient *http.Client) (*Constellation, error) {
	u.lock.RLock()
	constellationObj, ok := u.constellations[constellationID]
	if ok && !constellationObj.Expired() {
		u.lock.RUnlock()
		return constellationObj, nil
	}

	resp, err := eveonline.GetFromESI(
		fmt.Sprintf(ConstellationURLPattern, constellationID),
		httpClient,
		map[string][]string{},
	)
	if err != nil {
		u.lock.RUnlock()
		return nil, err
	}

	constellationObj = new(Constellation)
	err = json.Unmarshal(resp.Body, &constellationObj)
	if err != nil {
		u.lock.RUnlock()
		return nil, err
	}
	constellationObj.ExpiresAt = resp.ExpiresAt
	constellationObj.Etag = resp.Etag

	u.lock.RUnlock()
	u.lock.Lock()
	defer u.lock.Unlock()
	potentiallyRefreshedConstellationObj, ok := u.constellations[constellationID]
	if ok && !potentiallyRefreshedConstellationObj.Expired() {
		// Refreshed while upgrading the lock, so return the new version
		return potentiallyRefreshedConstellationObj, nil
	}

	u.constellations[constellationID] = constellationObj

	return constellationObj, nil
}
