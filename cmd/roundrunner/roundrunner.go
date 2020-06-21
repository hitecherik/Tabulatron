package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hitecherik/Imperial-Online-IV/internal/resolver"
	"github.com/hitecherik/Imperial-Online-IV/internal/roundrunner"
	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
	"github.com/hitecherik/Imperial-Online-IV/pkg/zoom"
	"github.com/joho/godotenv"
)

type options struct {
	round          uint
	csv            string
	db             string
	tabbycatApiKey string
	tabbycatUrl    string
	tabbycatSlug   string
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
	flag.UintVar(&opts.round, "round", 0, "the round to run")
	flag.StringVar(&opts.csv, "csv", "round.csv", "CSV file to allocate breakout rooms in")
	flag.StringVar(&opts.db, "db", "db.json", "JSON file to store zoom email information in")
	flag.BoolVar(&opts.verbose, "verbose", false, "print additional input")
	flag.Parse()

	if opts.round == 0 {
		fmt.Fprintln(os.Stderr, "please specify a round")
		os.Exit(2)
	}

	bail(godotenv.Load(envFile))

	opts.tabbycatApiKey = os.Getenv("TABBYCAT_API_KEY")
	opts.tabbycatUrl = os.Getenv("TABBYCAT_URL")
	opts.tabbycatSlug = os.Getenv("TABBYCAT_SLUG")
}

func main() {
	rawDatabase, err := ioutil.ReadFile(opts.db)
	bail(err)

	var database resolver.Database
	bail(json.Unmarshal(rawDatabase, &database))

	tabbycat := tabbycat.New(opts.tabbycatApiKey, opts.tabbycatUrl, opts.tabbycatSlug)
	rooms, err := tabbycat.GetRound(opts.round)
	bail(err)

	verbose("Fetched %v pairings\n", len(rooms))

	assignments, err := roundrunner.Allocate(database, rooms)
	bail(err)

	file, err := os.Create(opts.csv)
	bail(err)
	defer file.Close()

	leftover, err := zoom.WriteCsv(file, assignments)
	bail(err)

	verbose("%v assignments leftover\n", len(leftover))
}
