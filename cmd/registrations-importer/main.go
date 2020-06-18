package main

import (
	"fmt"
	"os"

	"github.com/hitecherik/Imperial-Online-IV/pkg/zoom"
	"github.com/joho/godotenv"
)

type credentials struct {
	zoomApiKey     string
	zoomApiSecret  string
	tabbycatApiKey string
}

var creds credentials
var meetingId string

func bail(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func init() {
	bail(godotenv.Load())

	creds.zoomApiKey = os.Getenv("ZOOM_API_KEY")
	creds.zoomApiSecret = os.Getenv("ZOOM_API_SECRET")
	creds.tabbycatApiKey = os.Getenv("TABBYCAT_API_KEY")
	meetingId = os.Getenv("ZOOM_MEETING_ID")
}

func main() {
	zoom := zoom.New(creds.zoomApiKey, creds.zoomApiSecret)

	registrants, err := zoom.GetRegistrants(meetingId)
	bail(err)

	fmt.Printf("Registrants:\n%+v\n", registrants)
}
