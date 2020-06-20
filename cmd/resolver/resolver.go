package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hitecherik/Imperial-Online-IV/internal/resolver"
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
	db             string
	verbose        bool
}

var opts options

func bail(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func verbose(format string, a ...interface{}) {
	if opts.verbose {
		fmt.Printf(format, a...)
	}
}

func init() {
	var envFile string

	flag.StringVar(&envFile, "env", ".env", "file to read environment variables from")
	flag.StringVar(&opts.db, "db", "db.json", "JSON file to store zoom email information in")
	flag.BoolVar(&opts.verbose, "verbose", false, "print additional input")
	flag.Parse()

	bail(godotenv.Load(envFile))

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

	verbose("Fetched %v registrants\n", len(registrants))

	tabbycat := tabbycat.New(opts.tabbycatApiKey, opts.tabbycatUrl, opts.tabbycatSlug)
	teams, err := tabbycat.GetTeams()
	bail(err)

	verbose("Fetched %v teams\n", len(teams))

	adjudicators, err := tabbycat.GetAdjudicators()
	bail(err)

	verbose("Fetched %v adjudicators\n", len(adjudicators))

	database := resolver.Resolve(registrants, teams, adjudicators)
	raw, err := json.Marshal(database)
	bail(err)
	bail(ioutil.WriteFile(opts.db, raw, 0644))
}
