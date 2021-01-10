package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/andersfylling/disgord"
	"github.com/hitecherik/Imperial-Online-IV/internal/db"
	"github.com/hitecherik/Imperial-Online-IV/internal/tabulatron"
	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
	"github.com/joho/godotenv"
)

func panic(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

type options struct {
	db              db.Database
	tabbycatApiKey  string
	tabbycatUrl     string
	tabbycatSlug    string
	botToken        string
	helperBotTokens []string
}

var opts options

func init() {
	var envFile string

	flag.StringVar(&envFile, "env", ".env", "file to read environment variables from")
	flag.Var(&opts.db, "db", "SQLite3 database representing the tournament")
	flag.Parse()

	panic(godotenv.Load(envFile))

	opts.tabbycatApiKey = os.Getenv("TABBYCAT_API_KEY")
	opts.tabbycatUrl = os.Getenv("TABBYCAT_URL")
	opts.tabbycatSlug = os.Getenv("TABBYCAT_SLUG")
	opts.botToken = os.Getenv("DISCORD_BOT_TOKEN")

	for i := 1; true; i++ {
		token := os.Getenv(fmt.Sprintf("DISCORD_HELPER_%v", i))

		if token == "" {
			break
		}

		opts.helperBotTokens = append(opts.helperBotTokens, token)
	}

	panic(opts.db.SetIfNotExists(fmt.Sprintf("%v.db", os.Getenv("TABBYCAT_SLUG"))))
}

func main() {
	for _, token := range opts.helperBotTokens {
		helperClient := disgord.New(disgord.Config{
			BotToken: token,
		})
		go helperClient.StayConnectedUntilInterrupted(context.Background())
	}

	client := disgord.New(disgord.Config{
		BotToken: opts.botToken,
	})
	defer client.StayConnectedUntilInterrupted(context.Background())

	tabbycat := tabbycat.New(opts.tabbycatApiKey, opts.tabbycatUrl, opts.tabbycatSlug)
	tron := tabulatron.New(client, &opts.db, tabbycat)

	me, err := client.Myself(context.Background())
	panic(err)

	guildId := 0

	client.On(disgord.EvtMessageCreate, func(s disgord.Session, evt *disgord.MessageCreate) {
		if me.ID == evt.Message.Author.ID {
			return
		}

		if guildId == 0 {
			guildId = int(evt.Message.GuildID)
			fmt.Printf("Bound to guild %v\n", evt.Message.GuildID)
		}

		if guildId != int(evt.Message.GuildID) || guildId == 0 {
			return
		}

		tron.HandleMessage(s, evt)
	})

	client.On(disgord.EvtGuildMemberRemove, func(s disgord.Session, evt *disgord.GuildMemberRemove) {
		if guildId == 0 || int(evt.GuildID) == 0 {
			return
		}

		tron.HandleDeparture(s, evt)
	})
}
