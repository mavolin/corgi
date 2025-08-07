package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

const entitiesURL = "https://html.spec.whatwg.org/entities.json"

type characterReference struct {
	Name  string
	Runes string
}

func loadCharacterReferences() ([]characterReference, error) {
	resp, err := http.Get(entitiesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch entities.json: %w", err)
	}

	entMap := make(map[string]struct {
		Characters string `json:"characters"`
	})

	err = json.NewDecoder(resp.Body).Decode(&entMap)
	if err != nil {
		return nil, fmt.Errorf("failed to decode entities.json: %w", err)
	}
	_ = resp.Body.Close()

	charRefs := make([]characterReference, 0, len(entMap))

	for name, obj := range entMap {
		if !strings.HasPrefix(name, "&") || !strings.HasSuffix(name, ";") {
			continue
		}
		name = name[len("&") : len(name)-len(";")]

		charRefs = append(charRefs, characterReference{
			Name:  name,
			Runes: obj.Characters,
		})
	}

	slices.SortFunc(charRefs, func(a, b characterReference) int {
		return strings.Compare(a.Name, b.Name)
	})

	return charRefs, nil
}
