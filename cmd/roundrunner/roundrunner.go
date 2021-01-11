package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hitecherik/Imperial-Online-IV/internal/db"
	"github.com/hitecherik/Imperial-Online-IV/internal/multiroom"
	"github.com/hitecherik/Imperial-Online-IV/internal/roundrunner"
	"github.com/hitecherik/Imperial-Online-IV/internal/rounds"
	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
	"github.com/hitecherik/Imperial-Online-IV/pkg/zoom"
	"github.com/joho/godotenv"
)

type options struct {
	round          rounds.Rounds
	db             db.Database
	tabbycatApiKey string
	tabbycatUrl    string
	tabbycatSlug   string
	categories     multiroom.Categories
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
	flag.Var(&opts.round, "round", "a round to run")
	flag.Var(&opts.db, "db", "SQLite3 database representing the tournament")
	flag.Var(&opts.categories, "categories", "path to the categories TOML document")
	flag.BoolVar(&opts.verbose, "verbose", false, "print additional input")
	flag.Parse()

	if len(opts.round) == 0 {
		fmt.Fprintln(os.Stderr, "please specify at least one round")
		os.Exit(2)
	}

	bail(godotenv.Load(envFile))

	opts.tabbycatApiKey = os.Getenv("TABBYCAT_API_KEY")
	opts.tabbycatUrl = os.Getenv("TABBYCAT_URL")
	opts.tabbycatSlug = os.Getenv("TABBYCAT_SLUG")

	bail(opts.db.SetIfNotExists(fmt.Sprintf("%v.db", opts.tabbycatSlug)))
}

func main() {
	var rooms []tabbycat.Room
	tabbycat := tabbycat.New(opts.tabbycatApiKey, opts.tabbycatUrl, opts.tabbycatSlug)

	for _, round := range opts.round {
		r, err := tabbycat.GetDraw(round)
		bail(err)
		rooms = append(rooms, r...)
	}

	verbose("Fetched %v pairings\n", len(rooms))

	venues, err := tabbycat.GetVenues()
	bail(err)

	verbose("Fetched %v venues\n", len(venues))

	assignments, err := roundrunner.Allocate(opts.db, venues, rooms, opts.categories)
	bail(err)

	written := 0

	for _, assignment := range assignments {
		if len(assignment.Allocation) == 0 {
			continue
		}

		base := fmt.Sprintf("round-%v", opts.round.String())

		if assignment.Category.Name != "" {
			base = fmt.Sprintf("%v-%v", base, strings.ToLower(assignment.Category.Name))
		}

		file, err := os.Create(fmt.Sprintf("%v.csv", base))
		bail(err)
		bail(zoom.WriteCsv(file, assignment.Allocation))
		file.Close()

		written += 1
	}

	verbose("Wrote %v files\n", written)
}
