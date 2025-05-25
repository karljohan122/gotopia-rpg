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
	Results []T `json:"results"`
}

type Monster struct {
	Name            string      `json:"name"`
	Description     Description `json:"desc"`
	HitPoints       int         `json:"hit_points"`
	ArmorClass      int         `json:"armor_class"`
	ChallengeRating string      `json:"challenge_rating"`
	Slug            string      `json:"slug"`
}

type Equipment struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Description string

func FetchRandomMonster() (*Monster, error) {
	rand.Seed(time.Now().UnixNano())
	for attempts := 0; attempts < 5; attempts++ {
		url := fmt.Sprintf("%s/monsters/?limit=100&offset=%d", BaseURL, rand.Intn(32)*100)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var monsters Paginated[Monster]
		if err := json.NewDecoder(resp.Body).Decode(&monsters); err != nil {
			return nil, err
		}

		rand.Shuffle(len(monsters.Results), func(i, j int) {
			monsters.Results[i], monsters.Results[j] = monsters.Results[j], monsters.Results[i]
		})

		for _, mon := range monsters.Results {
			desc := string(mon.Description)
			if desc != "" && desc != "False" {
				return &mon, nil
			}
		}
	}
	return nil, fmt.Errorf("no monsters with descriptions found after several attempts")
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

func (d *Description) UnmarshalJSON(data []byte) error {
	// If it's a string, unmarshal as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*d = Description(s)
		return nil
	}
	// If it's a boolean (false), treat as empty string
	var b bool
	if err := json.Unmarshal(data, &b); err == nil && b == false {
		*d = ""
		return nil
	}
	return fmt.Errorf("invalid description field: %s", string(data))
}
