package tabulatron

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/andersfylling/disgord"
)

const (
	registerRaw   string = `^!register(\d+)$`
	startregRaw   string = `^!startreg$`
	whitespaceRaw string = `\s`
)

var (
	register   *regexp.Regexp
	startreg   *regexp.Regexp
	whitespace *regexp.Regexp
)

type RegHandler struct {
	t              *Tabulatron
	regStarted     bool
	regChannel     *disgord.Channel
	regHelpChannel *disgord.Channel
	speakerRole    *disgord.Role
	judgeRole      *disgord.Role
	tabRole        *disgord.Role
}

func init() {
	var err error

	register, err = regexp.Compile(registerRaw)
	if err != nil {
		log.Fatal(err.Error())
	}

	startreg, err = regexp.Compile(startregRaw)
	if err != nil {
		log.Fatal(err.Error())
	}

	whitespace, err = regexp.Compile(whitespaceRaw)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func NewRegHandler(t *Tabulatron) *RegHandler {
	return &RegHandler{
		t:          t,
		regStarted: false,
	}
}

func (h *RegHandler) CanHandle(_ disgord.Session, evt *disgord.MessageCreate) bool {
	message := sanitiseMessage(evt.Message.Content)
	username := sanitiseMessage(evt.Message.Author.Username)

	return register.Match(message) || (len(message) == 0 && register.Match(username)) || startreg.Match(message)
}

func (h *RegHandler) Handle(s disgord.Session, evt *disgord.MessageCreate) {
	messageContent := sanitiseMessage(evt.Message.Content)
	usernameRegistration := len(messageContent) == 0

	if usernameRegistration {
		messageContent = sanitiseMessage(evt.Message.Author.Username)
	}

	if startreg.Match(messageContent) {
		if h.regStarted {
			h.t.ReplyMessage(evt.Message, "I can't do that. Registration has already started.")
			return
		}

		if err := h.populateChannels(evt); err != nil {
			log.Printf("error populating channels: %v", err.Error())
			return
		}

		if err := h.populateRoles(evt); err != nil {
			log.Printf("error populating roles: %v", err.Error())
			return
		}

		for _, role := range evt.Message.Member.Roles {
			if role == h.tabRole.ID {
				h.t.AcknowledgeMessage(s, evt.Message)
				h.regStarted = true
				return
			}
		}

		h.t.ReplyMessage(evt.Message, "you can't ask me to do that.")
		return
	}

	matches := register.FindSubmatch(messageContent)

	if len(matches) != 2 {
		log.Printf("unexpected number of matches in '%v'", string(messageContent))
		return
	}

	code := string(matches[1])

	if !h.regStarted {
		h.t.ReplyMessage(evt.Message, "I can't do that. Registration hasn't started yet.")
		return
	}

	if !usernameRegistration && evt.Message.ChannelID != h.regChannel.ID {
		h.t.ReplyMessage(
			evt.Message,
			"you can't do that here. Registration can only happen in the %v channel.",
			h.regChannel.Mention(),
		)
		return
	}

	if code == "123456" {
		h.t.ReplyMessage(
			evt.Message,
			"please replace `123456` in your message with your registration code. If you don't know what this is, ask in %v.",
			h.regHelpChannel.Mention(),
		)
		return
	}

	if len(code) != 6 {
		h.t.ReplyMessage(
			evt.Message,
			"please double-check your registration code – it should be six digits long.",
		)
		return
	}

	_, name, speaker, err := h.t.database.ParticipantFromBarcode(code, evt.Message.Author.ID.String())

	if err != nil {
		log.Printf("error registering speaker: %v", err.Error())
		h.t.ReplyMessage(evt.Message, "there was an error registering you. Please check the code you entered and try again.")
		return
	}

	h.t.ReplyMessage(evt.Message, "you have been successfully registered!")

	role := h.judgeRole
	if speaker {
		role = h.speakerRole
	}

	err = h.t.discord.
		UpdateGuildMember(context.Background(), evt.Message.GuildID, evt.Message.Author.ID).
		SetNick(name).
		SetRoles([]disgord.Snowflake{role.ID}).
		Execute()
	if err != nil {
		log.Printf("error setting nickname: %v", err.Error())
		h.t.ReplyMessage(
			evt.Message,
			"there was an error setting your nickname and/or role! Please ask in %v for help.",
			h.regHelpChannel.Mention(),
		)
	}
}

func (h *RegHandler) populateChannels(evt *disgord.MessageCreate) error {
	if h.regChannel != nil && h.regHelpChannel != nil {
		return nil
	}

	channels, err := h.t.discord.GetGuildChannels(context.Background(), evt.Message.GuildID)
	if err != nil {
		return err
	}

	for _, channel := range channels {
		if channel.Name == "registration-help" {
			h.regHelpChannel = channel
		} else if channel.Name == "registration" {
			h.regChannel = channel
		}
	}

	return nil
}

func (h *RegHandler) populateRoles(evt *disgord.MessageCreate) error {
	if h.speakerRole != nil && h.judgeRole != nil && h.tabRole != nil {
		return nil
	}

	roles, err := h.t.discord.GetGuildRoles(context.Background(), evt.Message.GuildID)
	if err != nil {
		return nil
	}

	for _, role := range roles {
		if role.Name == "Speaker" {
			h.speakerRole = role
		} else if role.Name == "Judge" {
			h.judgeRole = role
		} else if role.Name == "Tab/Tech" {
			h.tabRole = role
		}
	}

	return nil
}

func sanitiseMessage(message string) []byte {
	return whitespace.ReplaceAll([]byte(strings.ToLower(message)), []byte{})
}
