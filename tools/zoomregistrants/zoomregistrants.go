package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hitecherik/Tabulatron/pkg/zoom"
	"github.com/joho/godotenv"
)

type options struct {
	zoomApiKey     string
	zoomApiSecret  string
	zoomMeetingId  string
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
	flag.StringVar(&opts.db, "db", "registrants.json", "JSON file to store registrant information in")
	flag.BoolVar(&opts.verbose, "verbose", false, "print additional input")
	flag.Parse()

	bail(godotenv.Load(envFile))

	opts.zoomApiKey = os.Getenv("ZOOM_API_KEY")
	opts.zoomApiSecret = os.Getenv("ZOOM_API_SECRET")
	opts.zoomMeetingId = os.Getenv("ZOOM_MEETING_ID")
}

func main() {
	zoom := zoom.New(opts.zoomApiKey, opts.zoomApiSecret)

	registrants, err := zoom.GetRegistrants(opts.zoomMeetingId)
	bail(err)

	verbose("Fetched %v registrants\n", len(registrants))

	raw, err := json.Marshal(registrants)
	bail(err)
	bail(ioutil.WriteFile(opts.db, raw, 0644))
}
