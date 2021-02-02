package tabulatron

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/andersfylling/disgord"
)

const (
	checkinRaw      string = `^\s*[!1]\s*check[\s-]*in\s*$`
	startcheckinRaw string = `^\s*!startcheckin\s*$`
	endcheckinRaw   string = `^\s*!endcheckin\s*$`
)

var (
	checkin      *regexp.Regexp
	startcheckin *regexp.Regexp
	endcheckin   *regexp.Regexp
)

type CheckinHandler struct {
	t               *Tabulatron
	checkinStarted  bool
	checkinChannel  *disgord.Channel
	techHelpChannel *disgord.Channel
	tabRole         *disgord.Role
}

func init() {
	var err error

	checkin, err = regexp.Compile(checkinRaw)
	if err != nil {
		log.Fatal(err.Error())
	}

	startcheckin, err = regexp.Compile(startcheckinRaw)
	if err != nil {
		log.Fatal(err.Error())
	}

	endcheckin, err = regexp.Compile(endcheckinRaw)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func NewCheckinHandler(t *Tabulatron) *CheckinHandler {
	return &CheckinHandler{
		t:              t,
		checkinStarted: false,
	}
}

func (h *CheckinHandler) CanHandle(_ disgord.Session, evt *disgord.MessageCreate) bool {
	message := []byte(strings.ToLower(evt.Message.Content))

	return checkin.Match(message) || startcheckin.Match(message) || endcheckin.Match(message)
}

func (h *CheckinHandler) Handle(s disgord.Session, evt *disgord.MessageCreate) {
	rawMessage := []byte(strings.ToLower(evt.Message.Content))

	if startcheckin.Match(rawMessage) {
		if h.checkinStarted {
			h.t.ReplyMessage(evt.Message, "I can't do that. Check-in has already started.")
			h.t.RejectMessage(s, evt.Message)
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

		if h.hasTabRole(evt.Message.Member) {
			h.t.AcknowledgeMessage(s, evt.Message)
			h.checkinStarted = true
			return
		}

		h.t.ReplyMessage(evt.Message, "you can't ask me to do that.")
		h.t.RejectMessage(s, evt.Message)
		return
	}

	if !h.checkinStarted {
		h.t.ReplyMessage(evt.Message, "I can't do that. Check-in hasn't started yet.")
		h.t.RejectMessage(s, evt.Message)
		return
	}

	if endcheckin.Match(rawMessage) {
		if h.hasTabRole(evt.Message.Member) {
			h.t.AcknowledgeMessage(s, evt.Message)
			h.checkinStarted = false
			return
		}

		h.t.ReplyMessage(evt.Message, "you can't ask me to do that.")
		h.t.RejectMessage(s, evt.Message)
		return
	}

	if evt.Message.ChannelID != h.checkinChannel.ID {
		h.t.ReplyMessage(
			evt.Message,
			"you can't do that here. Check in can only happen in the %v channel.",
			h.checkinChannel.Mention(),
		)
		h.t.RejectMessage(s, evt.Message)
		return
	}

	id, speaker, err := h.t.database.ParticipantFromDiscord(evt.Message.Author.ID.String())
	if err != nil {
		log.Printf("error finding speaker: %v", err.Error())
		h.t.ReplyMessage(evt.Message, "there was an error checking you in. Please ask for help in %v.", h.techHelpChannel.Mention())
		h.t.RejectMessage(s, evt.Message)
		return
	}

	if err := h.t.tabbycat.CheckIn(id, speaker); err != nil {
		log.Printf("error checking in speaker: %v", err.Error())
		h.t.ReplyMessage(evt.Message, "there was an error checking you in. Please ask for help in %v.", h.techHelpChannel.Mention())
		h.t.RejectMessage(s, evt.Message)
		return
	}

	h.t.AcknowledgeMessage(s, evt.Message)
}

func (h *CheckinHandler) populateChannels(evt *disgord.MessageCreate) error {
	if h.checkinChannel != nil && h.techHelpChannel != nil {
		return nil
	}

	channels, err := h.t.discord.GetGuildChannels(context.Background(), evt.Message.GuildID)
	if err != nil {
		return err
	}

	for _, channel := range channels {
		if channel.Name == "tab-and-tech-help" {
			h.techHelpChannel = channel
		} else if channel.Name == "checkin" {
			h.checkinChannel = channel
		}
	}

	return nil
}

func (h *CheckinHandler) populateRoles(evt *disgord.MessageCreate) error {
	if h.tabRole != nil {
		return nil
	}

	roles, err := h.t.discord.GetGuildRoles(context.Background(), evt.Message.GuildID)
	if err != nil {
		return nil
	}

	for _, role := range roles {
		if role.Name == "Tab/Tech" {
			h.tabRole = role
		}
	}

	return nil
}

func (h *CheckinHandler) hasTabRole(member *disgord.Member) bool {
	for _, role := range member.Roles {
		if role == h.tabRole.ID {
			return true
		}
	}

	return false
}
