package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"

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

	opts.botTokens = []string{os.Getenv("DISCORD_BOT_TOKEN")}
	for i := 1; true; i++ {
		token := os.Getenv(fmt.Sprintf("DISCORD_HELPER_%v", i))

		if token == "" {
			break
		}

		opts.botTokens = append(opts.botTokens, token)
	}

	bail(opts.db.SetIfNotExists(fmt.Sprintf("%v.db", opts.tabbycatSlug)))
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

		for i, team := range room.TeamIds {
			discords, err := opts.db.DiscordFromTeamId(team)
			bail(err)

			snowflakes, err := stringsToSnowflakes(discords)
			bail(err)

			for _, snowflake := range snowflakes {
				bail(createDMAndSendMessage(
					clients[rand.Intn(len(clients))],
					snowflake,
					fmt.Sprintf("In this round, you will be speaking in **%v** in room **%v**.", room.SideNames[i], venueName),
				))
			}
		}

		bail(sendMessagesToJudges(clients, []string{room.ChairId}, "the chair", venueName))
		bail(sendMessagesToJudges(clients, room.PanellistIds, "a pannelist", venueName))
		bail(sendMessagesToJudges(clients, room.TraineeIds, "a trainee", venueName))

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

func sendMessagesToJudges(clients []*disgord.Client, ids []string, wingType string, venue string) error {
	discords, err := opts.db.DiscordFromParticipantIds(ids)
	if err != nil {
		return err
	}

	snowflakes, err := stringsToSnowflakes(discords)
	if err != nil {
		return err
	}

	for _, snowflake := range snowflakes {
		client := clients[rand.Intn(len(clients))]
		message := fmt.Sprintf("In this round, you will be judging as **%v** in room **%v**.", wingType, venue)

		if err := createDMAndSendMessage(client, snowflake, message); err != nil {
			return err
		}
	}

	return nil
}
