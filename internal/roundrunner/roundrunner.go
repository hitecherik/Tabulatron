package roundrunner

import (
	"fmt"

	"github.com/hitecherik/Imperial-Online-IV/internal/resolver"
	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
	"github.com/hitecherik/Imperial-Online-IV/pkg/zoom"
)

func Allocate(emails resolver.Database, venues []tabbycat.Venue, rooms []tabbycat.Room) [][]string {
	var allocations [][]string
	var panellists [][]string
	var trainees [][]string

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

		for _, trainee := range room.TraineeIds {
			if email, ok := emails.Judges[trainee]; ok {
				trainees = append(trainees, []string{name, email})
			}
		}
	}

	return append(allocations, append(panellists, trainees...)...)
}

func LeftoversToNames(leftovers [][]string, registrants []zoom.Registrant) [][]string {
	registrantMap := buildRegistrantMap(registrants)
	assignments := make([][]string, 0, len(leftovers))

	for _, leftover := range leftovers {
		if name, ok := registrantMap[leftover[1]]; ok {
			assignments = append(assignments, []string{leftover[0], name})
		} else {
			assignments = append(assignments, leftover)
		}
	}

	return assignments
}

func buildVenueMap(venues []tabbycat.Venue) map[string]string {
	venueMap := make(map[string]string)

	for _, venue := range venues {
		id := fmt.Sprintf("%v", venue.Id)
		venueMap[id] = venue.Name
	}

	return venueMap
}

func buildRegistrantMap(registrants []zoom.Registrant) map[string]string {
	registrantMap := make(map[string]string)

	for _, registrant := range registrants {
		registrantMap[registrant.Email] = registrant.Name
	}

	return registrantMap
}
