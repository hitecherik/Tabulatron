package tabbycat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

var identifierStripper *regexp.Regexp = regexp.MustCompile(`/(\d+)$`)

type Tabbycat struct {
	apiKey   string
	client   *http.Client
	endpoint string
}

type Team struct {
	Id       int64         `json:"id"`
	Speakers []Participant `json:"speakers"`
}

type Participant struct {
	Email string `json:"email"`
	Id    int64  `json:"id"`
	Name  string `json:"name"`
}

type Room struct {
	VenueId      string
	ChairId      string
	TeamIds      []string
	PanellistIds []string
}

type teamResponse struct {
	Adjudicators struct {
		Chair      string
		Panellists []string
		Trainees   []string
	}
	Teams []struct {
		Team string
	}
	Venue string
}

func New(apiKey string, url string, slug string) *Tabbycat {
	return &Tabbycat{
		apiKey:   apiKey,
		client:   &http.Client{},
		endpoint: fmt.Sprintf("%v/api/v1/tournaments/%v/", url, slug),
	}
}

func (t *Tabbycat) GetAdjudicators() ([]Participant, error) {
	req, err := http.NewRequest(http.MethodGet, t.endpoint+"adjudicators", nil)
	if err != nil {
		return nil, err
	}

	t.authorize(req)
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var adjudicators []Participant
	if err := json.Unmarshal(body, &adjudicators); err != nil {
		return nil, err
	}

	return adjudicators, nil
}

func (t *Tabbycat) GetTeams() ([]Team, error) {
	req, err := http.NewRequest(http.MethodGet, t.endpoint+"teams", nil)
	if err != nil {
		return nil, err
	}

	t.authorize(req)
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var teams []Team
	if err := json.Unmarshal(body, &teams); err != nil {
		return nil, err
	}

	return teams, nil
}

func (t *Tabbycat) GetRound(round uint) ([]Room, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%vrounds/%v/pairings", t.endpoint, round), nil)
	if err != nil {
		return nil, err
	}

	t.authorize(req)
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data []teamResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	rooms := make([]Room, 0, len(data))
	for _, datum := range data {
		venueId, err := stripIdentifier(datum.Venue)
		if err != nil {
			return nil, err
		}

		chairId, err := stripIdentifier(datum.Adjudicators.Chair)
		if err != nil {
			return nil, err
		}

		panellistIds := make([]string, 0, len(datum.Adjudicators.Panellists)+len(datum.Adjudicators.Trainees))
		for _, judge := range append(datum.Adjudicators.Panellists, datum.Adjudicators.Trainees...) {
			id, err := stripIdentifier(judge)
			if err != nil {
				return nil, err
			}

			panellistIds = append(panellistIds, id)
		}

		teamIds := make([]string, 0, len(datum.Teams))
		for _, team := range datum.Teams {
			id, err := stripIdentifier(team.Team)
			if err != nil {
				return nil, err
			}

			teamIds = append(teamIds, id)
		}

		rooms = append(rooms, Room{
			VenueId:      venueId,
			ChairId:      chairId,
			TeamIds:      teamIds,
			PanellistIds: panellistIds,
		})
	}

	return rooms, nil
}

func (t *Tabbycat) authorize(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Token %v", t.apiKey))
}

func stripIdentifier(url string) (string, error) {
	matches := identifierStripper.FindSubmatch([]byte(url))

	if len(matches) != 2 {
		return "", fmt.Errorf("could not parse identifier from url %v", url)
	}

	return string(matches[1]), nil
}
