package zoom

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

const expiryAfter time.Duration = 30 * time.Second
const endpoint string = "https://api.zoom.us/v2/meetings/%v/registrants"

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
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(endpoint, meetingId), nil)
	if err != nil {
		return nil, err
	}

	token, err := z.GetToken()
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token))

	resp, err := z.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data registrantResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	registrants := make([]Registrant, 0, len(data.Registrants))

	for _, registrant := range data.Registrants {
		registrants = append(registrants, Registrant{
			Email: registrant.Email,
			Name:  fmt.Sprintf("%v %v", registrant.FirstName, registrant.LastName),
		})
	}

	return registrants, nil
}
