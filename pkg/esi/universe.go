package esi

import (
	"encoding/json"
	"fmt"

	"github.com/pequalsnp/go-eveonline/pkg/eveonline"
)

type Type struct {
	ID        eveonline.TypeID  `json:"type_id"`
	GroupID   eveonline.GroupID `json:"group_id"`
	Volume    float64           `json:"volume"`
	Name      string            `json:"name"`
	Published bool              `json:"published"`
}

type Group struct {
	ID         eveonline.GroupID    `json:"group_id"`
	CategoryID eveonline.CategoryID `json:"category_id"`
	TypeIDs    []eveonline.TypeID   `json:"types"`
	Name       string               `json:"name"`
	Published  bool                 `json:"published"`
}

type Category struct {
	ID        eveonline.CategoryID `json:"category_id"`
	GroupIDs  []eveonline.GroupID  `json:"groups"`
	Name      string               `json:"name"`
	Published bool                 `json:"published"`
}

type Station struct {
	ID              eveonline.StationID       `json:"station_id"`
	Name            string                    `json:"name"`
	SystemID        eveonline.SystemID        `json:"system_id"`
	ConstellationID eveonline.ConstellationID `json:",omitempty"`
	RegionID        eveonline.RegionID        `json:",omitempty"`
}

type System struct {
	ID              eveonline.SystemID        `json:"system_id"`
	ConstellationID eveonline.ConstellationID `json:"constellation_id"`
}

type Constellation struct {
	ID       eveonline.ConstellationID `json:"constellation_id"`
	RegionID eveonline.RegionID        `json:"region_id"`
}

const TypeURLPattern = "https://esi.evetech.net/v3/universe/types/%d/"
const GroupURLPattern = "https://esi.evetech.net/v1/universe/groups/%d/"
const CategoryURLPattern = "https://esi.evetech.net/v1/universe/categories/%d/"
const StationURLPattern = "https://esi.evetech.net/v2/universe/stations/%d/"
const SystemURLPattern = "https://esi.evetech.net/v4/universe/systems/%d/"
const ConstellationURLPattern = "https://esi.evetech.net/v1/universe/constellations/%d/"

func (e *ESI) GetType(typeID eveonline.TypeID) (*Type, error) {
	resp, err := e.GetFromESI(
		fmt.Sprintf(TypeURLPattern, typeID),
		nil,
		map[string][]string{},
	)
	if err != nil {
		return nil, err
	}

	typeObj := new(Type)
	err = json.Unmarshal(resp.Body, &typeObj)
	if err != nil {
		return nil, err
	}

	return typeObj, nil
}

func (e *ESI) GetGroup(groupID eveonline.GroupID) (*Group, error) {
	resp, err := e.GetFromESI(
		fmt.Sprintf(GroupURLPattern, groupID),
		nil,
		map[string][]string{},
	)
	if err != nil {
		return nil, err
	}

	groupObj := new(Group)
	err = json.Unmarshal(resp.Body, &groupObj)
	if err != nil {
		return nil, err
	}

	return groupObj, nil
}

func (e *ESI) GetCategory(categoryID eveonline.CategoryID) (*Category, error) {
	resp, err := e.GetFromESI(
		fmt.Sprintf(CategoryURLPattern, categoryID),
		nil,
		map[string][]string{},
	)
	if err != nil {
		return nil, err
	}

	categoryObj := new(Category)
	err = json.Unmarshal(resp.Body, &categoryObj)
	if err != nil {
		return nil, err
	}

	return categoryObj, nil
}

func (e *ESI) GetStation(stationID eveonline.StationID) (*Station, error) {
	resp, err := e.GetFromESI(
		fmt.Sprintf(StationURLPattern, stationID),
		nil,
		map[string][]string{},
	)
	if err != nil {
		return nil, err
	}

	stationObj := new(Station)
	err = json.Unmarshal(resp.Body, &stationObj)
	if err != nil {
		return nil, err
	}

	system, err := e.GetSystem(stationObj.SystemID)
	if err != nil {
		return nil, err
	}

	constellation, err := e.GetConstellation(system.ConstellationID)
	if err != nil {
		return nil, err
	}

	stationObj.ConstellationID = constellation.ID
	stationObj.RegionID = constellation.RegionID

	return stationObj, nil
}

func (e *ESI) GetSystem(systemID eveonline.SystemID) (*System, error) {
	resp, err := e.GetFromESI(
		fmt.Sprintf(SystemURLPattern, systemID),
		nil,
		map[string][]string{},
	)
	if err != nil {
		return nil, err
	}

	systemObj := new(System)
	err = json.Unmarshal(resp.Body, systemObj)
	if err != nil {
		return nil, err
	}

	return systemObj, nil
}

func (e *ESI) GetConstellation(constellationID eveonline.ConstellationID) (*Constellation, error) {
	resp, err := e.GetFromESI(
		fmt.Sprintf(ConstellationURLPattern, constellationID),
		nil,
		map[string][]string{},
	)
	if err != nil {
		return nil, err
	}

	constellationObj := new(Constellation)
	err = json.Unmarshal(resp.Body, &constellationObj)
	if err != nil {
		return nil, err
	}

	return constellationObj, nil
}
