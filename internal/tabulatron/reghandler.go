package tabulatron

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/andersfylling/disgord"
)

const (
	registerRaw       string = `^[1!]?register[^\d]*(\d+)$`
	numbersRaw        string = `^\d{6}$`
	startregRaw       string = `^!startreg$`
	linkRaw           string = `!link<@!\d+>(\d{6})$`
	whitespaceRaw     string = `\s`
	maxNicknameLength int    = 32
)

var (
	register   *regexp.Regexp
	numbers    *regexp.Regexp
	startreg   *regexp.Regexp
	link       *regexp.Regexp
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

func bail(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func init() {
	var err error

	register, err = regexp.Compile(registerRaw)
	bail(err)

	numbers, err = regexp.Compile(numbersRaw)
	bail(err)

	startreg, err = regexp.Compile(startregRaw)
	bail(err)

	link, err = regexp.Compile(linkRaw)
	bail(err)

	whitespace, err = regexp.Compile(whitespaceRaw)
	bail(err)
}

func NewRegHandler(t *Tabulatron) *RegHandler {
	return &RegHandler{
		t:          t,
		regStarted: false,
	}
}

func (h *RegHandler) CanHandle(_ disgord.Session, evt *disgord.MessageCreate) bool {
	message := sanitiseMessage(evt.Message.Content)

	if err := h.populateChannels(evt); err != nil {
		log.Printf("error populating channels: %v", err.Error())
	} else if evt.Message.ChannelID == h.regChannel.ID && numbers.Match(message) {
		return true
	}

	if evt.Message.Type == disgord.MessageTypeGuildMemberJoin {
		username := sanitiseMessage(evt.Message.Author.Username)
		return register.Match(username) || numbers.Match(username)
	}

	return register.Match(message) || startreg.Match(message) || link.Match(message)
}

func (h *RegHandler) Handle(s disgord.Session, evt *disgord.MessageCreate) {
	messageContent := sanitiseMessage(evt.Message.Content)
	usernameRegistration := evt.Message.Type == disgord.MessageTypeGuildMemberJoin
	linkRegistration := link.Match(messageContent)

	if usernameRegistration {
		messageContent = sanitiseMessage(evt.Message.Author.Username)
	}

	author := evt.Message.Author
	if linkRegistration {
		author = evt.Message.Mentions[0]
	}

	if startreg.Match(messageContent) {
		if h.regStarted {
			h.t.ReplyMessage(evt.Message, "I can't do that. Registration has already started.")
			h.t.RejectMessage(s, evt.Message)
			return
		}

		if err := h.populateRoles(evt); err != nil {
			log.Printf("error populating roles: %v", err.Error())
			return
		}

		if h.hasTabRole(evt.Message.Member) {
			h.t.AcknowledgeMessage(s, evt.Message)
			h.regStarted = true
			return
		}

		h.t.ReplyMessage(evt.Message, "you can't ask me to do that.")
		h.t.RejectMessage(s, evt.Message)
		return
	}

	if linkRegistration && !h.hasTabRole(evt.Message.Member) {
		h.t.ReplyMessage(evt.Message, "you can't ask me to do that.")
		h.t.RejectMessage(s, evt.Message)
		return
	}

	re := register
	if linkRegistration {
		re = link
	}

	code := string(messageContent)
	matches := re.FindSubmatch(messageContent)

	if matches != nil {
		if len(matches) != 2 {
			log.Printf("unexpected number of matches in '%v'", string(messageContent))
			return
		}

		code = string(matches[1])
	}

	if !h.regStarted {
		h.t.ReplyMessage(evt.Message, "I can't do that. Registration hasn't started yet.")
		h.t.RejectMessage(s, evt.Message)
		return
	}

	if !usernameRegistration && evt.Message.ChannelID != h.regChannel.ID {
		h.t.ReplyMessage(
			evt.Message,
			"you can't do that here. Registration can only happen in the %v channel.",
			h.regChannel.Mention(),
		)
		h.t.RejectMessage(s, evt.Message)
		return
	}

	if code == "123456" {
		h.t.ReplyMessage(
			evt.Message,
			"please replace `123456` in your message with your registration code. If you don't know what this is, ask in %v.",
			h.regHelpChannel.Mention(),
		)
		h.t.RejectMessage(s, evt.Message)
		return
	}

	if len(code) != 6 {
		h.t.ReplyMessage(
			evt.Message,
			"please double-check your registration code – it should be six digits long.",
		)
		h.t.RejectMessage(s, evt.Message)
		return
	}

	_, name, speaker, err := h.t.database.ParticipantFromBarcode(code, author.ID.String())

	if err != nil {
		log.Printf("error registering speaker: %v", err.Error())
		h.t.ReplyMessage(evt.Message, "there was an error registering you. Please check the code you entered and try again.")
		h.t.RejectMessage(s, evt.Message)
		return
	}

	h.t.AcknowledgeMessage(s, evt.Message)

	role := h.judgeRole
	if speaker {
		role = h.speakerRole
	}

	if len(name) > maxNicknameLength {
		name = name[:maxNicknameLength]
	}

	err = h.t.discord.
		UpdateGuildMember(context.Background(), evt.Message.GuildID, author.ID).
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

	go h.t.CreateDMAndSendMessage(author.ID, "Congratulations! You have successfully registered.")
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

func (h *RegHandler) hasTabRole(member *disgord.Member) bool {
	for _, role := range member.Roles {
		if role == h.tabRole.ID {
			return true
		}
	}

	return false
}

func sanitiseMessage(message string) []byte {
	return whitespace.ReplaceAll([]byte(strings.ToLower(message)), []byte{})
}
