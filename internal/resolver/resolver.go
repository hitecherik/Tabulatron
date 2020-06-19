package resolver

import (
	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
	"github.com/hitecherik/Imperial-Online-IV/pkg/zoom"
)

type Database struct {
	Resolved map[string][]string `json:"resolved"`
	Unknown  []zoom.Registrant   `json:"unknown"`
}

func Resolve(registrants *[]zoom.Registrant, teams *[]tabbycat.Team, adjudicators *[]tabbycat.Participant) Database {
	return Database{map[string][]string{}, *registrants}
}
