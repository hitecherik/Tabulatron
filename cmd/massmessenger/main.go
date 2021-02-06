package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/andersfylling/disgord"
	"github.com/hitecherik/Tabulatron/internal/db"
	"github.com/hitecherik/Tabulatron/internal/hermes"
	"github.com/joho/godotenv"
)

type options struct {
	db           db.Database
	botTokens    []string
	tabbycatSlug string
	verbose      bool
	message      string
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
	flag.Var(&opts.db, "db", "SQLite3 database representing the tournament")
	flag.BoolVar(&opts.verbose, "verbose", false, "print additional output")
	flag.StringVar(&opts.message, "message", "", "the message to send all participants")
	flag.Parse()

	bail(godotenv.Load(envFile))

	if opts.message == "" {
		bail(fmt.Errorf("Please provide a non-empty message to send all participants."))
	}

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
	clients := make([]*hermes.Hermes, 0, len(opts.botTokens))
	for _, token := range opts.botTokens {
		client := disgord.New(disgord.Config{
			BotToken: token,
		})
		go client.StayConnectedUntilInterrupted(context.Background())

		h := hermes.New(client)
		clients = append(clients, h)

		go h.Listen()
		defer h.Wait()
	}

	discords, err := opts.db.AllDiscords()
	bail(err)

	snowflakes, err := stringsToSnowflakes(discords)
	bail(err)

	for i, snowflake := range snowflakes {
		clients[i%len(clients)].SendMessage(snowflake, opts.message)
	}

	verbose("Queued %v messages.\n", len(snowflakes))
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
