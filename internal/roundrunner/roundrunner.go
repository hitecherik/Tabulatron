package roundrunner

import (
	"fmt"

	"github.com/hitecherik/Imperial-Online-IV/internal/db"
	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
)

func Allocate(database db.Database, venues []tabbycat.Venue, rooms []tabbycat.Room) ([][]string, error) {
	var allocations [][]string
	var panellists [][]string
	var trainees [][]string

	venueMap := BuildVenueMap(venues)

	for _, room := range rooms {
		name := venueMap[room.VenueId]

		if emails, err := database.ParticipantEmails([]string{room.ChairId}); err == nil && len(emails) != 0 {
			allocations = append(allocations, []string{name, emails[0]})
		} else if err != nil {
			return nil, err
		}

		if emails, err := database.TeamEmails(room.TeamIds); err == nil && len(emails) != 0 {
			allocations = appendEmails(allocations, name, emails)
		} else if err != nil {
			return nil, err
		}

		if emails, err := database.ParticipantEmails(room.PanellistIds); err == nil && len(emails) != 0 {
			panellists = appendEmails(panellists, name, emails)
		} else if err != nil {
			return nil, err
		}

		if emails, err := database.ParticipantEmails(room.TraineeIds); err == nil && len(emails) != 0 {
			trainees = appendEmails(trainees, name, emails)
		} else if err != nil {
			return nil, err
		}
	}

	return append(allocations, append(panellists, trainees...)...), nil
}

func LeftoversToNames(database db.Database, leftovers [][]string) ([][]string, error) {
	assignments := make([][]string, 0, len(leftovers))

	for _, leftover := range leftovers {
		name, err := database.ParticipantNameFromEmail(leftover[1])
		if err != nil {
			return nil, err
		}

		assignments = append(assignments, []string{leftover[0], name})
	}

	return assignments, nil
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
