package main

import (
	"fmt"
	"os"

	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
	"github.com/hitecherik/Imperial-Online-IV/pkg/zoom"
	"github.com/joho/godotenv"
)

type options struct {
	zoomApiKey     string
	zoomApiSecret  string
	zoomMeetingId  string
	tabbycatApiKey string
	tabbycatUrl    string
	tabbycatSlug   string
}

var opts options

func bail(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func init() {
	bail(godotenv.Load())

	opts.zoomApiKey = os.Getenv("ZOOM_API_KEY")
	opts.zoomApiSecret = os.Getenv("ZOOM_API_SECRET")
	opts.zoomMeetingId = os.Getenv("ZOOM_MEETING_ID")

	opts.tabbycatApiKey = os.Getenv("TABBYCAT_API_KEY")
	opts.tabbycatUrl = os.Getenv("TABBYCAT_URL")
	opts.tabbycatSlug = os.Getenv("TABBYCAT_SLUG")
}

func main() {
	zoom := zoom.New(opts.zoomApiKey, opts.zoomApiSecret)

	registrants, err := zoom.GetRegistrants(opts.zoomMeetingId)
	bail(err)

	fmt.Printf("Registrants:\n%+v\n\n", registrants)

	tabbycat := tabbycat.New(opts.tabbycatApiKey, opts.tabbycatUrl, opts.tabbycatSlug)
	teams, err := tabbycat.GetTeams()
	bail(err)

	fmt.Printf("Teams:\n%+v\n\n", teams)

	adjudicators, err := tabbycat.GetAdjudicators()
	bail(err)

	fmt.Printf("Adjudicators:\n%+v\n", adjudicators)
}
