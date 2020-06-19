package resolver

import (
	"fmt"
	"strings"

	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
	"github.com/hitecherik/Imperial-Online-IV/pkg/zoom"
)

type Database struct {
	Resolved map[string][]string `json:"resolved"`
	Unknown  []zoom.Registrant   `json:"unknown"`
}

func Resolve(registrants []zoom.Registrant, teams []tabbycat.Team, adjudicators []tabbycat.Participant) Database {
	database := Database{map[string][]string{}, []zoom.Registrant{}}

outer:
	for _, registrant := range registrants {
		for _, team := range teams {
			id := fmt.Sprintf("%v", team.Id)

			for _, speaker := range team.Speakers {
				if compare(&registrant, &speaker) {
					if _, ok := database.Resolved[id]; !ok {
						database.Resolved[id] = []string{}
					}

					database.Resolved[id] = append(database.Resolved[id], registrant.Email)
					continue outer
				}
			}
		}

		for _, judge := range adjudicators {
			if compare(&registrant, &judge) {
				id := fmt.Sprintf("%v", judge.Id)
				database.Resolved[id] = []string{registrant.Email}
				continue outer
			}
		}

		database.Unknown = append(database.Unknown, registrant)
	}

	return database
}

func compare(registrant *zoom.Registrant, participant *tabbycat.Participant) bool {
	if strings.ToLower(registrant.Email) == strings.ToLower(participant.Email) {
		return true
	}

	if strings.ToLower(registrant.Name) == strings.ToLower(participant.Name) {
		return true
	}

	return false
}
