package esiutil

import (
	"github.com/pequalsnp/go-eveonline/pkg/esi"
	"github.com/pequalsnp/go-eveonline/pkg/eveonline"
)

type result struct {
	types []*esi.Type
	err   error
}

func AllTypesForGroup(e *esi.ESI, groupID eveonline.GroupID) ([]*esi.Type, error) {
	group, err := e.GetGroup(groupID)
	if err != nil {
		return nil, err
	}

	var types []*esi.Type
	for _, typeID := range group.TypeIDs {
		typeObj, err := e.GetType(typeID)
		if err != nil {
			return nil, err
		}
		types = append(types, typeObj)
	}

	return types, nil
}

func AllTypesForCategory(e *esi.ESI, categoryID eveonline.CategoryID) ([]*esi.Type, error) {
	category, err := e.GetCategory(categoryID)
	if err != nil {
		return nil, err
	}

	var types []*esi.Type
	typesChan := make(chan result)
	for _, groupID := range category.GroupIDs {
		go func(gid eveonline.GroupID) {
			typesForGroup, err := AllTypesForGroup(e, gid)
			typesChan <- result{typesForGroup, err}
		}(groupID)
	}

	for range category.GroupIDs {
		groupResult := <-typesChan
		if groupResult.err != nil {
			return nil, groupResult.err
		}
		types = append(types, groupResult.types...)
	}

	return types, nil
}
