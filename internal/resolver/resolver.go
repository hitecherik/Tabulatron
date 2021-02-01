package resolver

import (
	"fmt"
	"strings"

	"github.com/hitecherik/Tabulatron/pkg/tabbycat"
	"github.com/hitecherik/Tabulatron/pkg/zoom"
)

type Database struct {
	Teams   map[string][]string `json:"teams"`
	Judges  map[string]string   `json:"judges"`
	Unknown []zoom.Registrant   `json:"unknown"`
}

func Resolve(registrants []zoom.Registrant, teams []tabbycat.Team, adjudicators []tabbycat.Participant) Database {
	database := Database{map[string][]string{}, map[string]string{}, []zoom.Registrant{}}

outer:
	for _, registrant := range registrants {
		for _, team := range teams {
			id := fmt.Sprintf("%v", team.Id)

			for _, speaker := range team.Speakers {
				if compare(&registrant, &speaker) {
					if _, ok := database.Teams[id]; !ok {
						database.Teams[id] = []string{}
					}

					database.Teams[id] = append(database.Teams[id], registrant.Email)
					continue outer
				}
			}
		}

		for _, judge := range adjudicators {
			if compare(&registrant, &judge) {
				id := fmt.Sprintf("%v", judge.Id)
				database.Judges[id] = registrant.Email
				continue outer
			}
		}

		database.Unknown = append(database.Unknown, registrant)
	}

	return database
}

func compare(registrant *zoom.Registrant, participant *tabbycat.Participant) bool {
	if strings.EqualFold(registrant.Email, participant.Email) {
		return true
	}

	if strings.EqualFold(registrant.Name, participant.Name) {
		return true
	}

	return false
}
