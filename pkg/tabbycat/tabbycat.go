package tabbycat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var identifierStripper *regexp.Regexp = regexp.MustCompile(`/(\d+)$`)

type Tabbycat struct {
	apiKey      string
	client      *http.Client
	endpoint    string
	privateUrls string
}

type Team struct {
	Id       uint          `json:"id"`
	Emoji    string        `json:"emoji"`
	Speakers []Participant `json:"speakers"`
}

type Participant struct {
	Email   string `json:"email"`
	Id      uint   `json:"id"`
	Name    string `json:"name"`
	Barcode string
	UrlKey  string `json:"url_key"`
}

type Room struct {
	VenueId      string
	ChairId      string
	TeamIds      []string
	PanellistIds []string
	TraineeIds   []string
	SideNames    []string
}

type Round struct {
	Id     string
	Name   string
	Motion Motion
}

type Venue struct {
	Id   uint
	Name string
}

type Motion struct {
	InfoSlide string `json:"info_slide"`
	Motion    string `json:"text"`
}

type teamResponse struct {
	Adjudicators struct {
		Chair      string
		Panellists []string
		Trainees   []string
	}
	Teams []struct {
		Side string
		Team string
	}
	Venue string
}

type roundResponse struct {
	Url     string
	Name    string
	Motions []Motion
}

func New(apiKey string, url string, slug string) *Tabbycat {
	return &Tabbycat{
		apiKey:      apiKey,
		client:      &http.Client{},
		endpoint:    fmt.Sprintf("%v/api/v1/tournaments/%v/", url, slug),
		privateUrls: fmt.Sprintf("%v/%v/privateurls/", url, slug),
	}
}

func (t *Tabbycat) GetAdjudicators() ([]Participant, error) {
	response, err := t.makeRequest(http.MethodGet, "adjudicators", nil)
	if err != nil {
		return nil, err
	}

	var adjudicators []Participant
	if err := json.Unmarshal(response, &adjudicators); err != nil {
		return nil, err
	}

	if err := t.GetBarcodes(false, adjudicators); err != nil {
		return nil, err
	}

	return adjudicators, nil
}

func (t *Tabbycat) GetTeams() ([]Team, error) {
	response, err := t.makeRequest(http.MethodGet, "teams", nil)
	if err != nil {
		return nil, err
	}

	var teams []Team
	if err := json.Unmarshal(response, &teams); err != nil {
		return nil, err
	}

	for i := range teams {
		if err := t.GetBarcodes(true, teams[i].Speakers); err != nil {
			return nil, err
		}
	}

	return teams, nil
}

func (t *Tabbycat) GetBarcodes(speakers bool, participants []Participant) error {
	category := "adjudicators"
	if speakers {
		category = "speakers"
	}

	for i := range participants {
		raw, err := t.makeRequest(http.MethodGet, fmt.Sprintf("%v/%v/checkin", category, participants[i].Id), nil)
		if err != nil {
			return err
		}

		var barcode struct{ Barcode string }
		if err := json.Unmarshal(raw, &barcode); err != nil {
			return err
		}

		participants[i].Barcode = barcode.Barcode
	}

	return nil
}

func (t *Tabbycat) GetRounds() ([]Round, error) {
	response, err := t.makeRequest(http.MethodGet, "rounds", nil)
	if err != nil {
		return nil, err
	}

	var responses []roundResponse
	if err := json.Unmarshal(response, &responses); err != nil {
		return nil, err
	}

	rounds := make([]Round, 0, len(responses))
	for _, response := range responses {
		round, err := response.toRound()
		if err != nil {
			return nil, err
		}

		rounds = append(rounds, round)
	}

	return rounds, nil
}

func (t *Tabbycat) GetRound(round uint64) (Round, error) {
	response, err := t.makeRequest(http.MethodGet, fmt.Sprintf("rounds/%v", round), nil)
	if err != nil {
		return Round{}, err
	}

	var rawRound roundResponse
	if err := json.Unmarshal(response, &rawRound); err != nil {
		return Round{}, err
	}

	return rawRound.toRound()
}

func (t *Tabbycat) ReleaseMotion(round uint64, starts time.Time) error {
	path := fmt.Sprintf("rounds/%v", round)
	response, err := t.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}

	var roundObject map[string]interface{}
	if err := json.Unmarshal(response, &roundObject); err != nil {
		return err
	}

	roundObject["motions_released"] = true
	roundObject["starts_at"] = starts.Format("15:04:05")

	serialized, err := json.Marshal(roundObject)
	if err != nil {
		return err
	}

	_, err = t.makeRequest(http.MethodPost, path, bytes.NewReader(serialized))
	return err
}

func (t *Tabbycat) GetDraw(round uint64) ([]Room, error) {
	response, err := t.makeRequest(http.MethodGet, fmt.Sprintf("rounds/%v/pairings", round), nil)
	if err != nil {
		return nil, err
	}

	var data []teamResponse
	if err := json.Unmarshal(response, &data); err != nil {
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

		panellistIds, err := stripIdentifiers(datum.Adjudicators.Panellists)
		if err != nil {
			return nil, err
		}

		traineeIds, err := stripIdentifiers(datum.Adjudicators.Trainees)
		if err != nil {
			return nil, err
		}

		teamIds := make([]string, 0, len(datum.Teams))
		sideNames := make([]string, 0, len(datum.Teams))
		for _, team := range datum.Teams {
			id, err := stripIdentifier(team.Team)
			if err != nil {
				return nil, err
			}

			teamIds = append(teamIds, id)
			sideNames = append(sideNames, strings.ToUpper(team.Side))
		}

		rooms = append(rooms, Room{
			VenueId:      venueId,
			ChairId:      chairId,
			TeamIds:      teamIds,
			PanellistIds: panellistIds,
			TraineeIds:   traineeIds,
			SideNames:    sideNames,
		})
	}

	return rooms, nil
}

func (t *Tabbycat) GetVenues() ([]Venue, error) {
	response, err := t.makeRequest(http.MethodGet, "venues", nil)
	if err != nil {
		return nil, err
	}

	var venues []Venue
	if err := json.Unmarshal(response, &venues); err != nil {
		return nil, err
	}

	return venues, nil
}

func (t *Tabbycat) CheckOutAdjudicator(id uint) error {
	_, err := t.makeRequest(http.MethodDelete, fmt.Sprintf("adjudicators/%v/checkin", id), nil)
	return err
}

func (t *Tabbycat) CheckIn(id uint, speaker bool) error {
	category := "adjudicators"
	if speaker {
		category = "speakers"
	}

	_, err := t.makeRequest(http.MethodPut, fmt.Sprintf("%v/%v/checkin", category, id), nil)
	return err
}

func (t *Tabbycat) PrivateUrlFromKey(urlKey string) string {
	return fmt.Sprintf("%v%v/", t.privateUrls, urlKey)
}

func (t *Tabbycat) makeRequest(method string, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, t.endpoint+url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %v", t.apiKey))

	if method == http.MethodPost {
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func stripIdentifier(url string) (string, error) {
	matches := identifierStripper.FindSubmatch([]byte(url))

	if len(matches) != 2 {
		return "", fmt.Errorf("could not parse identifier from url %v", url)
	}

	return string(matches[1]), nil
}

func stripIdentifiers(urls []string) ([]string, error) {
	ids := make([]string, 0, len(urls))

	for _, url := range urls {
		id, err := stripIdentifier(url)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func (r *roundResponse) toRound() (Round, error) {
	id, err := stripIdentifier(r.Url)
	if err != nil {
		return Round{}, err
	}

	var motion Motion
	if len(r.Motions) > 0 {
		motion = r.Motions[0]
	}

	return Round{id, r.Name, motion}, nil
}
