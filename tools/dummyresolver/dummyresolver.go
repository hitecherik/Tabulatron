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

	opts.tabbycatApiKey = os.Getenv("TABBYCAT_API_KEY")
	opts.tabbycatUrl = os.Getenv("TABBYCAT_URL")
	opts.tabbycatSlug = os.Getenv("TABBYCAT_SLUG")
}

func main() {
	tabbycat := tabbycat.New(opts.tabbycatApiKey, opts.tabbycatUrl, opts.tabbycatSlug)
	teams, err := tabbycat.GetTeams()
	bail(err)

	verbose("Fetched %v teams\n", len(teams))

	adjudicators, err := tabbycat.GetAdjudicators()
	bail(err)

	verbose("Fetched %v adjudicators\n", len(adjudicators))

	database := resolver.Database{
		Teams:   map[string][]string{},
		Judges:  map[string]string{},
		Unknown: make([]zoom.Registrant, 0),
	}

	for _, team := range teams {
		id := fmt.Sprintf("%v", team.Id)
		database.Teams[id] = make([]string, 0, len(team.Speakers))

		for _, speaker := range team.Speakers {
			database.Teams[id] = append(database.Teams[id], speaker.Email)
		}
	}

	for _, judge := range adjudicators {
		database.Judges[fmt.Sprintf("%v", judge.Id)] = judge.Email
	}

	raw, err := json.Marshal(database)
	bail(err)
	bail(ioutil.WriteFile(opts.db, raw, 0644))
}
