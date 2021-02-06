package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hitecherik/Tabulatron/internal/db"
	"github.com/hitecherik/Tabulatron/pkg/tabbycat"
	"github.com/joho/godotenv"
)

var opts struct {
	tabbycatApiKey string
	tabbycatUrl    string
	tabbycatSlug   string
	verbose        bool
	redact         bool
	reset          bool
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
	flag.BoolVar(&opts.redact, "redact", false, "redact participants' names")
	flag.BoolVar(&opts.verbose, "verbose", false, "print additional output")
	flag.BoolVar(&opts.reset, "reset", false, "whether to wipe the database")
	flag.Var(&opts.db, "db", "SQLite3 database representing the tournament")
	flag.Parse()

	bail(godotenv.Load(envFile))

	opts.tabbycatApiKey = os.Getenv("TABBYCAT_API_KEY")
	opts.tabbycatUrl = os.Getenv("TABBYCAT_URL")
	opts.tabbycatSlug = os.Getenv("TABBYCAT_SLUG")

	bail(opts.db.SetIfNotExists(fmt.Sprintf("%v.db", opts.tabbycatSlug)))
}

func main() {
	if opts.reset {
		bail(opts.db.Reset())
	}

	tabbycat := tabbycat.New(opts.tabbycatApiKey, opts.tabbycatUrl, opts.tabbycatSlug)
	teams, err := tabbycat.GetTeams()
	bail(err)

	verbose("Fetched %v teams\n", len(teams))

	if opts.redact {
		for i := range teams {
			redactNames(teams[i].Speakers)
		}
	}

	adjudicators, err := tabbycat.GetAdjudicators()
	bail(err)

	if opts.redact {
		redactNames(adjudicators)
	}

	verbose("Fetched %v adjudicators\n", len(adjudicators))

	bail(opts.db.AddTeams(teams))
	verbose("Inserted %v teams into database\n", len(teams))

	bail(opts.db.AddParticipants(false, adjudicators))
	verbose("Inserted %v adjudicators into database\n", len(adjudicators))
}

func redactNames(participants []tabbycat.Participant) {
	for i := range participants {
		components := strings.Split(participants[i].Name, " ")
		redacted := make([]string, 1, len(components))
		redacted[0] = components[0]

		for _, component := range components[1:] {
			redacted = append(redacted, strings.ToUpper(string(component[0])))
		}

		participants[i].Name = strings.Join(redacted, " ")
	}
}
