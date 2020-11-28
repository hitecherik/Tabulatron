package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/andersfylling/disgord"
	"github.com/hitecherik/Imperial-Online-IV/internal/db"
	"github.com/hitecherik/Imperial-Online-IV/internal/roundrunner"
	"github.com/hitecherik/Imperial-Online-IV/internal/rounds"
	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
	"github.com/joho/godotenv"
)

type options struct {
	round          rounds.Rounds
	db             db.Database
	botTokens      []string
	tabbycatApiKey string
	tabbycatUrl    string
	tabbycatSlug   string
	urlsPath       string
	verbose        bool
}

type url struct {
	Prefix string
	Url    string
}

var opts options
var urls []url

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
	flag.BoolVar(&opts.verbose, "verbose", false, "print additional input")
	flag.StringVar(&opts.urlsPath, "urls", "", "path to the zoom URLs json document")
	flag.Parse()

	if len(opts.round) == 0 {
		fmt.Fprintln(os.Stderr, "please specify at least one round")
		os.Exit(2)
	}

	bail(godotenv.Load(envFile))

	opts.tabbycatApiKey = os.Getenv("TABBYCAT_API_KEY")
	opts.tabbycatUrl = os.Getenv("TABBYCAT_URL")
	opts.tabbycatSlug = os.Getenv("TABBYCAT_SLUG")

	opts.botTokens = []string{os.Getenv("DISCORD_BOT_TOKEN")}
	for i := 1; true; i++ {
		token := os.Getenv(fmt.Sprintf("DISCORD_HELPER_%v", i))

		if token == "" {
			break
		}

		opts.botTokens = append(opts.botTokens, token)
	}

	bail(opts.db.SetIfNotExists(fmt.Sprintf("%v.db", opts.tabbycatSlug)))

	if opts.urlsPath == "" {
		urls = []url{}
	} else {
		urlsRaw, err := ioutil.ReadFile(opts.urlsPath)
		bail(err)
		bail(json.Unmarshal(urlsRaw, &urls))
	}
}

func main() {
	clients := make([]*disgord.Client, 0, len(opts.botTokens))
	for _, token := range opts.botTokens {
		client := disgord.New(disgord.Config{
			BotToken: token,
		})
		defer client.StayConnectedUntilInterrupted(context.Background())

		clients = append(clients, client)
	}

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
	venueMap := roundrunner.BuildVenueMap(venues)

	verbose("Fetched %v venues\n", len(venues))

	for _, room := range rooms {
		venueName := venueMap[room.VenueId]
		venueUrl := ""

		for _, url := range urls {
			if strings.HasPrefix(venueName, url.Prefix) {
				venueUrl = url.Url
				break
			}
		}

		for i, team := range room.TeamIds {
			discords, urlKeys, err := opts.db.ParticipantsFromTeamId(team)
			bail(err)

			snowflakes, err := stringsToSnowflakes(discords)
			bail(err)

			for j, snowflake := range snowflakes {
				privateUrl := tabbycat.PrivateUrlFromKey(urlKeys[j])
				message := fmt.Sprintf("In this round, you will be speaking in **%v** in room **%v**.", room.SideNames[i], venueName)

				if venueUrl != "" {
					message = fmt.Sprintf("%v\n\nThe link to your Zoom room is %v.", message, venueUrl)
				}

				if privateUrl != "" {
					message = fmt.Sprintf("%v\n\nYour private URL is %v.", message, privateUrl)
				}

				if err := createDMAndSendMessage(clients[rand.Intn(len(clients))], snowflake, message); err != nil {
					log.Printf("Error sending message to %v: %v", snowflake, err.Error())
				}
			}
		}

		bail(sendMessagesToJudges(clients, tabbycat, []string{room.ChairId}, "the chair", venueName, venueUrl))
		bail(sendMessagesToJudges(clients, tabbycat, room.PanellistIds, "a panellist", venueName, venueUrl))
		bail(sendMessagesToJudges(clients, tabbycat, room.TraineeIds, "a trainee", venueName, venueUrl))

		verbose("Sent messages for room %v\n", venueName)
	}

	os.Exit(0)
}

func stringToSnowflake(str string) (disgord.Snowflake, error) {
	snowflake, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}

	return disgord.NewSnowflake(snowflake), nil
}

func stringsToSnowflakes(strs []string) ([]disgord.Snowflake, error) {
	snowflakes := make([]disgord.Snowflake, 0, len(strs))
	for _, discord := range strs {
		snowflake, err := stringToSnowflake(discord)
		if err != nil {
			return nil, err
		}

		snowflakes = append(snowflakes, snowflake)
	}

	return snowflakes, nil
}

func createDMAndSendMessage(client *disgord.Client, snowflake disgord.Snowflake, message string) error {
	channel, err := client.CreateDM(context.Background(), snowflake)
	if err != nil {
		return err
	}

	_, err = channel.SendMsgString(context.Background(), client, message)
	return err
}

func sendMessagesToJudges(clients []*disgord.Client, tabbycat *tabbycat.Tabbycat, ids []string, wingType string, venue string, url string) error {
	discords, urlKeys, err := opts.db.DiscordFromParticipantIds(ids)
	if err != nil {
		return err
	}

	snowflakes, err := stringsToSnowflakes(discords)
	if err != nil {
		return err
	}

	for i, snowflake := range snowflakes {
		client := clients[rand.Intn(len(clients))]
		message := fmt.Sprintf("In this round, you will be judging as **%v** in room **%v**.", wingType, venue)

		if url != "" {
			message = fmt.Sprintf("%v\n\nThe link to your Zoom room is %v.", message, url)
		}

		if urlKeys[i] != "" {
			message = fmt.Sprintf("%v\n\nYour private URL is %v.", message, tabbycat.PrivateUrlFromKey(urlKeys[i]))
		}

		if err := createDMAndSendMessage(client, snowflake, message); err != nil {
			log.Printf("Error sending message to %v: %v", snowflake, err.Error())
		}
	}

	return nil
}
