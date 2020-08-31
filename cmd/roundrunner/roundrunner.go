package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hitecherik/Imperial-Online-IV/internal/db"
	"github.com/hitecherik/Imperial-Online-IV/internal/roundrunner"
	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
	"github.com/hitecherik/Imperial-Online-IV/pkg/zoom"
	"github.com/joho/godotenv"
	"github.com/olekukonko/tablewriter"
)

type rounds []uint64

type options struct {
	round          rounds
	csv            string
	db             db.Database
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
	flag.Var(&opts.round, "round", "a round to run")
	flag.StringVar(&opts.csv, "csv", "round.csv", "CSV file to allocate breakout rooms in")
	flag.Var(&opts.db, "db", "SQLite3 database representing the tournament")
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
		r, err := tabbycat.GetRound(round)
		bail(err)
		rooms = append(rooms, r...)
	}

	verbose("Fetched %v pairings\n", len(rooms))

	venues, err := tabbycat.GetVenues()
	bail(err)

	verbose("Fetched %v venues\n", len(venues))

	assignments, err := roundrunner.Allocate(opts.db, venues, rooms)
	bail(err)

	file, err := os.Create(opts.csv)
	bail(err)
	defer file.Close()

	leftovers, err := zoom.WriteCsv(file, assignments)
	bail(err)

	if len(leftovers) > 0 {
		verbose("%v assignments leftover\n", len(leftovers))

		assignments, err := roundrunner.LeftoversToNames(opts.db, leftovers)
		bail(err)

		fmt.Println("Please manually perform the following breakout room assignments:")

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Room"})

		for _, assignment := range assignments {
			table.Append(assignment)
		}

		table.Render()
	}
}

func (rs *rounds) String() string {
	ids := make([]string, 0, len(*rs))

	for _, r := range *rs {
		ids = append(ids, fmt.Sprintf("%v", r))
	}

	return strings.Join(ids, ",")
}

func (r *rounds) Set(s string) error {
	round, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}

	*r = append(*r, round)
	return nil
}
