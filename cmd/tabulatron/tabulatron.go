package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/andersfylling/disgord"
	"github.com/hitecherik/Imperial-Online-IV/internal/db"
	"github.com/joho/godotenv"
)

func panic(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

const register = `^\s*!register\s*(\d{6})\s*$`
const startreg = `^\s*!startreg\s*$`
const speakerRoleName = "Speaker"
const judgeRoleName = "Judge"
const regChannelName = "registration"
const helpChannelName = "registration-help"

var data db.Database
var botToken string

func init() {
	var envFile string

	flag.StringVar(&envFile, "env", ".env", "file to read environment variables from")
	flag.Var(&data, "db", "SQLite3 database representing the tournament")
	flag.Parse()

	panic(godotenv.Load(envFile))
	panic(data.SetIfNotExists(fmt.Sprintf("%v.db", os.Getenv("TABBYCAT_SLUG"))))
	botToken = os.Getenv("DISCORD_BOT_TOKEN")
}

func main() {
	client := disgord.New(disgord.Config{
		BotToken: botToken,
	})
	defer client.StayConnectedUntilInterrupted(context.Background())

	me, err := client.Myself(context.Background())
	panic(err)

	register, err := regexp.Compile(register)
	panic(err)

	startreg, err := regexp.Compile(startreg)
	panic(err)

	var (
		speakerRole disgord.Snowflake
		judgeRole   disgord.Snowflake
		helpChannel string
		regChannel  *disgord.Channel
	)

	client.On(disgord.EvtMessageCreate, func(s disgord.Session, evt *disgord.MessageCreate) {
		if me.ID == evt.Message.Author.ID {
			return
		}

		if startreg.Match([]byte(evt.Message.Content)) {
			if err := evt.Message.React(context.Background(), s, "👍"); err != nil {
				log.Printf("error reacting: %v\n", err.Error())
			}

			channels, err := client.GetGuildChannels(context.Background(), evt.Message.GuildID)
			if err != nil {
				log.Printf("error getting channels: %v", err.Error())
			}

			for _, channel := range channels {
				if channel.Name == helpChannelName {
					helpChannel = channel.Mention()
				} else if channel.Name == regChannelName {
					regChannel = channel
				}
			}

			roles, err := client.GetGuildRoles(context.Background(), evt.Message.GuildID)
			if err != nil {
				log.Printf("error getting roles: %v", err.Error())
				return
			}

			for _, role := range roles {
				if role.Name == speakerRoleName {
					speakerRole = role.ID
				} else if role.Name == judgeRoleName {
					judgeRole = role.ID
				}
			}

			return
		}

		matches := register.FindSubmatch([]byte(evt.Message.Content))

		if len(matches) != 2 {
			return
		}

		code := string(matches[1])

		var message string
		if regChannel == nil {
			message = fmt.Sprintf("%v, registration is not currently open.", evt.Message.Author.Mention())
		} else if evt.Message.ChannelID != regChannel.ID {
			message = fmt.Sprintf(
				"%v, registration can only happen in the %v channel.",
				evt.Message.Author.Mention(),
				regChannel.Mention(),
			)
		} else if code == "123456" {
			message = fmt.Sprintf(
				"%v, please replace `123456` in your message with your registration code. If you don't know what this is, ask in %v.",
				evt.Message.Author.Mention(),
				helpChannel,
			)
		} else {
			_, name, speaker, err := data.ParticipantFromBarcode(code, evt.Message.Author.ID.String())
			if err != nil {
				message = fmt.Sprintf("%v, there was an error registering you!", evt.Message.Author.Mention())
				log.Printf("error registering speaker: %v", err.Error())
			} else {
				message = fmt.Sprintf("%v, you have been successfuly registered!", evt.Message.Author.Mention())

				role := judgeRole
				if speaker {
					role = speakerRole
				}

				err = client.
					UpdateGuildMember(context.Background(), evt.Message.GuildID, evt.Message.Author.ID).
					SetNick(name).
					SetRoles([]disgord.Snowflake{role}).
					Execute()
				if err != nil {
					log.Printf("error setting nickname: %v", err.Error())
					message = fmt.Sprintf(
						"%v, there was an error setting your nickname and/or role! Please ask in %v.",
						evt.Message.Author.Mention(),
						helpChannel,
					)
				}
			}

		}

		_, err = evt.Message.Reply(context.Background(), s, message)
		if err != nil {
			log.Printf("could not send reply: %v", err.Error())
		}
	})
}
