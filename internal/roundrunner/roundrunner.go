package roundrunner

import (
	"fmt"
	"log"

	"github.com/hitecherik/Imperial-Online-IV/internal/db"
	"github.com/hitecherik/Imperial-Online-IV/internal/multiroom"
	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
)

type Allocation struct {
	Category   multiroom.Category
	Allocation [][]string
}

func Allocate(database db.Database, venues []tabbycat.Venue, rooms []tabbycat.Room, categories multiroom.Categories) ([]Allocation, error) {
	allocations := make(map[string]*Allocation)

	for _, category := range categories {
		allocations[category.Name] = &Allocation{Category: category}
	}

	allocations[""] = &Allocation{Category: multiroom.Category{}}

	venueMap := BuildVenueMap(venues)

	for _, room := range rooms {
		name := venueMap[room.VenueId]
		category, err := categories.Lookup(name)
		if err != nil {
			log.Print(err.Error())
		}

		judgeIds := append(append([]string{room.ChairId}, room.PanellistIds...), room.TraineeIds...)

		if emails, err := database.TeamEmails(room.TeamIds); err == nil && len(emails) != 0 {
			allocations[category.Name].Allocation = appendEmails(allocations[category.Name].Allocation, name, emails)
		} else if err != nil {
			return nil, err
		}

		if emails, err := database.ParticipantEmails(judgeIds); err == nil && len(emails) != 0 {
			allocations[category.Name].Allocation = appendEmails(allocations[category.Name].Allocation, name, emails)
		} else if err != nil {
			return nil, err
		}
	}

	values := make([]Allocation, 0, len(allocations))

	for _, allocation := range allocations {
		values = append(values, *allocation)
	}

	return values, nil
}

func BuildVenueMap(venues []tabbycat.Venue) map[string]string {
	venueMap := make(map[string]string)

	for _, venue := range venues {
		id := fmt.Sprintf("%v", venue.Id)
		venueMap[id] = venue.Name
	}

	return venueMap
}

func appendEmails(allocations [][]string, name string, emails []string) [][]string {
	for _, email := range emails {
		allocations = append(allocations, []string{name, email})
	}

	return allocations
}
