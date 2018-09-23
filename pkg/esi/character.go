package esi

import (
	"encoding/json"
	"fmt"
	"net/http"

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

const CharacterDetailsURLPattern = "https://esi.evetech.net/v4/characters/%d/"
const CharacterPortraitsURLPattern = "https://esi.evetech.net/v2/characters/%d/portrait"
const CharacterSkillsURLPattern = "https://esi.evetech.net/v4/characters/%d/skills"

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
