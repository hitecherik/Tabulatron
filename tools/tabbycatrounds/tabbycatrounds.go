package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
	"github.com/joho/godotenv"
	"github.com/olekukonko/tablewriter"
)

type options struct {
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
	flag.BoolVar(&opts.verbose, "verbose", false, "print additional input")
	flag.Parse()

	bail(godotenv.Load(envFile))

	opts.tabbycatApiKey = os.Getenv("TABBYCAT_API_KEY")
	opts.tabbycatUrl = os.Getenv("TABBYCAT_URL")
	opts.tabbycatSlug = os.Getenv("TABBYCAT_SLUG")
}

func main() {
	tabbycat := tabbycat.New(opts.tabbycatApiKey, opts.tabbycatUrl, opts.tabbycatSlug)
	rounds, err := tabbycat.GetRounds()
	bail(err)

	verbose("Fetched %v rounds\n", len(rounds))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name"})

	for _, round := range rounds {
		table.Append([]string{fmt.Sprintf("%v", round.Id), round.Name})
	}

	table.Render()
}
