package roundrunner

import (
	"fmt"

	"github.com/hitecherik/Imperial-Online-IV/internal/resolver"
	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
)

func Allocate(emails resolver.Database, venues []tabbycat.Venue, rooms []tabbycat.Room) ([][]string, error) {
	var allocations [][]string
	var panellists [][]string

	venueMap := buildVenueMap(venues)

	for _, room := range rooms {
		name := venueMap[room.VenueId]

		if email, ok := emails.Judges[room.ChairId]; ok {
			allocations = append(allocations, []string{name, email})
		}

		for _, team := range room.TeamIds {
			if speakers, ok := emails.Teams[team]; ok {
				for _, speaker := range speakers {
					allocations = append(allocations, []string{name, speaker})
				}
			}
		}

		for _, panellist := range room.PanellistIds {
			if email, ok := emails.Judges[panellist]; ok {
				panellists = append(panellists, []string{name, email})
			}
		}
	}

	return append(allocations, panellists...), nil
}

func buildVenueMap(venues []tabbycat.Venue) map[string]string {
	venueMap := make(map[string]string)

	for _, venue := range venues {
		id := fmt.Sprintf("%v", venue.Id)
		venueMap[id] = venue.Name
	}

	return venueMap
}
