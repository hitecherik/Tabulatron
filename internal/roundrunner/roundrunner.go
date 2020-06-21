package roundrunner

import (
	"fmt"

	"github.com/hitecherik/Imperial-Online-IV/internal/resolver"
	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
)

func Allocate(emails resolver.Database, rooms []tabbycat.Room) ([][]string, error) {
	var allocations [][]string
	var panellists [][]string

	for _, room := range rooms {
		// TODO: actually get the room name
		name := fmt.Sprintf("venue%v", room.VenueId)

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
