package zoom

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

const expiryAfter time.Duration = 30 * time.Second
const endpoint string = "https://api.zoom.us/v2/meetings/%v/registrants"
const maxPageSize int = 300

type Zoom struct {
	apiKey    string
	apiSecret string
	client    *http.Client
}

type Registrant struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type registrantResponse struct {
	Registrants []struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string
	}
	PageCount int `json:"page_count"`
}

func New(apiKey string, apiSecret string) *Zoom {
	return &Zoom{apiKey, apiSecret, &http.Client{}}
}

func (z *Zoom) GetToken() (string, error) {
	claims := jwt.Claims{
		Issuer: z.apiKey,
		Expiry: jwt.NewNumericDate(time.Now().Add(expiryAfter).UTC()),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	signer, err := jose.NewSigner(jose.SigningKey{
		Algorithm: jose.HS256,
		Key:       []byte(z.apiSecret),
	}, nil)
	if err != nil {
		return "", err
	}

	token, err := signer.Sign(payload)
	if err != nil {
		return "", err
	}

	return token.CompactSerialize()
}

func (z *Zoom) GetRegistrants(meetingId string) ([]Registrant, error) {
	data, err := z.retrievePaginatedRegistrants(meetingId)
	if err != nil {
		return nil, err
	}

	registrants := make([]Registrant, 0, len(data.Registrants))

	for _, registrant := range data.Registrants {
		nameComponents := make([]string, 0, 2)

		if registrant.FirstName != "" {
			nameComponents = append(nameComponents, registrant.FirstName)
		}

		if registrant.LastName != "" {
			nameComponents = append(nameComponents, registrant.LastName)
		}

		registrants = append(registrants, Registrant{
			Email: registrant.Email,
			Name:  strings.Join(nameComponents, " "),
		})
	}

	return registrants, nil
}

func (z *Zoom) retrievePaginatedRegistrants(meetingId string) (registrantResponse, error) {
	pageCount := 1
	responses := make([]registrantResponse, 0, pageCount)

	for pageNumber := 1; pageNumber <= pageCount; pageNumber++ {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(endpoint, meetingId), nil)
		if err != nil {
			return registrantResponse{}, err
		}

		token, err := z.GetToken()
		if err != nil {
			return registrantResponse{}, err
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token))

		q := req.URL.Query()
		q.Add("page_size", fmt.Sprintf("%v", maxPageSize))
		q.Add("page_number", fmt.Sprintf("%v", pageNumber))
		req.URL.RawQuery = q.Encode()

		resp, err := z.client.Do(req)
		if err != nil {
			return registrantResponse{}, err
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return registrantResponse{}, err
		}

		var data registrantResponse
		if err := json.Unmarshal(body, &data); err != nil {
			return registrantResponse{}, err
		}

		pageCount = data.PageCount
		responses = append(responses, data)
	}

	response := responses[0]
	for _, data := range responses[1:] {
		response.Registrants = append(response.Registrants, data.Registrants...)
	}

	return response, nil
}
