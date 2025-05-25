package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const BaseURL = "https://api.open5e.com"

type Paginated[T any] struct {
	Count   int `json:"count"`
	Results []T `json:"results"`
}

type Monster struct {
	Name            string   `json:"name"`
	Description     string   `json:"desc"`
	HitPoints       int      `json:"hit_points"`
	ArmorClass      int      `json:"armor_class"`
	ChallengeRating string   `json:"challenge_rating"`
	Slug            string   `json:"slug"`
	Actions         []Action `json:"actions"`
}

type Action struct {
	Name        string `json:"name"`
	Desc        string `json:"desc"`
	AttackBonus int    `json:"attack_bonus"`
	DamageDice  string `json:"damage_dice"`
}

type Equipment struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func FetchRandomMonster() (*Monster, error) {
	rand.Seed(time.Now().UnixNano())

	metaURL := fmt.Sprintf("%s/monsters/?challenge_rating=1&limit=1", BaseURL)
	metaResp, err := http.Get(metaURL)
	if err != nil {
		return nil, err
	}
	defer metaResp.Body.Close()

	var meta Paginated[Monster]
	if err := json.NewDecoder(metaResp.Body).Decode(&meta); err != nil {
		return nil, err
	}
	if meta.Count == 0 {
		return nil, fmt.Errorf("no CR 1 monsters found")
	}

	pageSize := 20
	pageCount := (meta.Count + pageSize - 1) / pageSize

	for attempts := 0; attempts < 5; attempts++ {
		page := rand.Intn(pageCount) + 1
		url := fmt.Sprintf("%s/monsters/?challenge_rating=1&page=%d", BaseURL, page)

		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var mons Paginated[Monster]
		if err := json.NewDecoder(resp.Body).Decode(&mons); err != nil {
			return nil, err
		}

		rand.Shuffle(len(mons.Results), func(i, j int) {
			mons.Results[i], mons.Results[j] = mons.Results[j], mons.Results[i]
		})

		for _, m := range mons.Results {
			d := m.Description
			if d != "" && d != "False" {
				return &m, nil
			}
		}
	}

	return nil, fmt.Errorf("no CR 1 monster with description found after several attempts")
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
