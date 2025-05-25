package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
)

const BaseURL = "https://api.open5e.com"

type Paginated[T any] struct {
	Results []T `json:"results"`
}

type Monster struct {
	Name            string `json:"name"`
	Description     string `json:"desc"`
	HitPoints       int    `json:"hit_points"`
	ArmorClass      int    `json:"armor_class"`
	ChallengeRating string `json:"challenge_rating"`
	Slug            string `json:"slug"`
}

type Equipment struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func FetchRandomMonster() (*Monster, error) {
	url := fmt.Sprintf("%s/monsters", BaseURL)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var monsters Paginated[Monster]
	if err := json.NewDecoder(resp.Body).Decode(&monsters); err != nil {
		return nil, err
	}

	if len(monsters.Results) == 0 {
		return nil, fmt.Errorf("no monsters found")
	}

	i := rand.Intn(len(monsters.Results))
	return &monsters.Results[i], nil
}

func FetchRandomEquipment() (*Equipment, error) {
	url := fmt.Sprintf("%s/equipment", BaseURL)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var equipment Paginated[Equipment]
	if err := json.NewDecoder(resp.Body).Decode(&equipment); err != nil {
		return nil, err
	}

	if len(equipment.Results) == 0 {
		return nil, fmt.Errorf("no equipment found")
	}

	i := rand.Intn(len(equipment.Results))
	return &equipment.Results[i], nil
}
