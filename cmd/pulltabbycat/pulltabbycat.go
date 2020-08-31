package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hitecherik/Imperial-Online-IV/internal/db"
	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
	"github.com/joho/godotenv"
)

var opts struct {
	tabbycatApiKey string
	tabbycatUrl    string
	tabbycatSlug   string
	verbose        bool
	db             db.Database
}

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
	flag.BoolVar(&opts.verbose, "verbose", false, "print additional input")
	flag.Var(&opts.db, "db", "SQLite3 database representing the tournament")
	flag.Parse()

	bail(godotenv.Load(envFile))

	opts.tabbycatApiKey = os.Getenv("TABBYCAT_API_KEY")
	opts.tabbycatUrl = os.Getenv("TABBYCAT_URL")
	opts.tabbycatSlug = os.Getenv("TABBYCAT_SLUG")

	bail(opts.db.SetIfNotExists(fmt.Sprintf("%v.db", opts.tabbycatSlug)))
}

func main() {
	bail(opts.db.Reset())

	tabbycat := tabbycat.New(opts.tabbycatApiKey, opts.tabbycatUrl, opts.tabbycatSlug)
	teams, err := tabbycat.GetTeams()
	bail(err)

	verbose("Fetched %v teams\n", len(teams))

	adjudicators, err := tabbycat.GetAdjudicators()
	bail(err)

	verbose("Fetched %v adjudicators\n", len(adjudicators))

	bail(opts.db.AddTeams(teams))
	verbose("Inserted %v teams into database\n", len(teams))

	bail(opts.db.AddParticipants(false, adjudicators))
	verbose("Inserted %v adjudicators into database\n", len(adjudicators))
}