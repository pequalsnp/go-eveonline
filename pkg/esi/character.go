package esi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pequalsnp/go-eveonline/pkg/eveonline"
)

type Portraits struct {
	Px128Url string `json:"px128x128"`
	Px256URL string `json:"px256x256"`
	Px512URL string `json:"px512x512"`
	Px64URL  string `json:"px64x64"`
}

type Character struct {
	ID        eveonline.CharacterID `json:"id,omitempty"`
	Name      string                `json:"name"`
	Portraits *Portraits            `json:"portraits,omitempty"`
}

type Skill struct {
	ID           eveonline.SkillID `json:"skill_id"`
	TrainedLevel int               `json:"trained_skill_level"`
	ActiveLevel  int               `json:"active_skill_level"`
	Skillpoints  int64             `json:"skillpoints_in_skill"`
}

type characterSkillsESI struct {
	Skills []*Skill
}

type CharacterSkills struct {
	Skills map[eveonline.SkillID]*Skill
}

type Asset struct {
	IsBlueprintCopy bool                 `json:"is_blueprint_copy"`
	IsSingleton     bool                 `json:""is_singleton`
	LocationID      eveonline.LocationID `json:"location_id"`
	LocationType    string               `json:"location_type"`
	TypeID          eveonline.TypeID     `json:"type_id"`
	Quantity        int64                `json:"quantity"`
}

type CharacterAssets struct {
	Assets []*Asset
}

const CharacterDetailsURLPattern = "https://esi.evetech.net/v4/characters/%d/"
const CharacterPortraitsURLPattern = "https://esi.evetech.net/v2/characters/%d/portrait"
const CharacterSkillsURLPattern = "https://esi.evetech.net/v4/characters/%d/skills"
const CharacterAssetsURLPattern = "https://esi.evetech.net/v3/characters/%d/assets/"
const CharacterWalletBalanceURLPattern = "https://esi.evetech.net/v1/characters/%d/wallet/"

func GetCharacterDetails(httpClient *http.Client, characterID eveonline.CharacterID) (*Character, error) {
	url := fmt.Sprintf(CharacterDetailsURLPattern, characterID)
	resp, err := eveonline.GetFromESI(url, httpClient, map[string][]string{})
	if err != nil {
		return nil, err
	}

	character := new(Character)
	err = json.Unmarshal(resp.Body, character)
	if err != nil {
		return nil, err
	}
	character.ID = characterID

	url = fmt.Sprintf(CharacterPortraitsURLPattern, characterID)
	resp, err = eveonline.GetFromESI(url, httpClient, map[string][]string{})
	if err != nil {
		return nil, err
	}
	portraits := new(Portraits)
	err = json.Unmarshal(resp.Body, portraits)
	if err != nil {
		return nil, err
	}
	character.Portraits = portraits

	return character, nil
}

func GetCharacterSkills(httpClient *http.Client, characterID eveonline.CharacterID) (*CharacterSkills, error) {
	url := fmt.Sprintf(CharacterSkillsURLPattern, characterID)
	resp, err := eveonline.GetFromESI(url, httpClient, map[string][]string{})
	if err != nil {
		return nil, err
	}

	esiSkills := new(characterSkillsESI)
	err = json.Unmarshal(resp.Body, &esiSkills)
	if err != nil {
		return nil, err
	}

	skills := make(map[eveonline.SkillID]*Skill)
	for _, skill := range esiSkills.Skills {
		skills[skill.ID] = skill
	}
	return &CharacterSkills{Skills: skills}, nil
}

func GetCharacterAssets(authdClient *http.Client, characterID eveonline.CharacterID) (*CharacterAssets, error) {
	characterAssetsURL := fmt.Sprintf(CharacterAssetsURLPattern, characterID)

	allPages, err := eveonline.GetAllPages(characterAssetsURL, 1, authdClient)
	if err != nil {
		return nil, fmt.Errorf("Failed to get assets for character id %d, %v", characterID, err)
	}

	var latestExpiry time.Time
	assets := make([]*Asset, 0)
	for _, page := range allPages {
		if page.ExpiresAt.After(latestExpiry) {
			latestExpiry = page.ExpiresAt
		}

		err = json.Unmarshal(page.Body, &assets)
		if err != nil {
			return nil, fmt.Errorf("Failed while unmarshalling assets for character %d, %v", characterID, err)
		}
	}

	return &CharacterAssets{Assets: assets}, nil
}

func GetCharacterWalletBalance(authdClient *http.Client, characterID eveonline.CharacterID) (float64, error) {
	characterWalletURL := fmt.Sprintf(CharacterWalletBalanceURLPattern, characterID)

	resp, err := eveonline.GetFromESI(characterWalletURL, authdClient, map[string][]string{})
	if err != nil {
		return 0.0, err
	}

	balance, err := strconv.ParseFloat(string(resp.Body), 64)
	if err != nil {
		return 0.0, fmt.Errorf("Failed to parse wallet balance for character id %d", characterID)
	}

	return balance, nil
}
