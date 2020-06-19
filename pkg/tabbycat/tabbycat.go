package tabbycat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

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

func (t *Tabbycat) authorize(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Token %v", t.apiKey))
}
