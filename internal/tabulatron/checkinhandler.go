package tabulatron

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/andersfylling/disgord"
)

var (
	checkin      *regexp.Regexp = regexp.MustCompile(`^\s*[!1]\s*ch[ie]ck[\s-]*[ie]n\s*$`)
	chicken      *regexp.Regexp = regexp.MustCompile(`chicken`)
	checkout     *regexp.Regexp = regexp.MustCompile(`^\s*[!1]\s*check[\s-]*out\s*$`)
	startcheckin *regexp.Regexp = regexp.MustCompile(`^\s*!startcheckin\s*$`)
	endcheckin   *regexp.Regexp = regexp.MustCompile(`^\s*!endcheckin\s*$`)
)

type CheckinHandler struct {
	t               *Tabulatron
	checkinStarted  bool
	checkinChannel  *disgord.Channel
	checkoutChannel *disgord.Channel
	techHelpChannel *disgord.Channel
	tabRole         *disgord.Role
	judgeRole       *disgord.Role
}

func NewCheckinHandler(t *Tabulatron) *CheckinHandler {
	return &CheckinHandler{
		t:              t,
		checkinStarted: false,
	}
}

func (h *CheckinHandler) CanHandle(_ disgord.Session, evt *disgord.MessageCreate) bool {
	message := []byte(strings.ToLower(evt.Message.Content))

	return checkin.Match(message) || checkout.Match(message) || startcheckin.Match(message) || endcheckin.Match(message)
}

func (h *CheckinHandler) Handle(s disgord.Session, evt *disgord.MessageCreate) {
	rawMessage := []byte(strings.ToLower(evt.Message.Content))

	if startcheckin.Match(rawMessage) {
		if h.checkinStarted {
			h.t.ReplyMessage(evt.Message, "I can't do that. Check-in has already started.")
			h.t.RejectMessage(evt.Message)
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
			h.t.AcknowledgeMessage(evt.Message)
			h.checkinStarted = true
			return
		}

		h.t.ReplyMessage(evt.Message, "you can't ask me to do that.")
		h.t.RejectMessage(evt.Message)
		return
	}

	if !h.checkinStarted {
		h.t.ReplyMessage(evt.Message, "I can't do that. Check-in hasn't started yet.")
		h.t.RejectMessage(evt.Message)
		return
	}

	if endcheckin.Match(rawMessage) {
		if h.hasTabRole(evt.Message.Member) {
			h.t.AcknowledgeMessage(evt.Message)
			h.checkinStarted = false
			return
		}

		h.t.ReplyMessage(evt.Message, "you can't ask me to do that.")
		h.t.RejectMessage(evt.Message)
		return
	}

	if (evt.Message.ChannelID != h.checkinChannel.ID && evt.Message.ChannelID != h.checkoutChannel.ID) ||
		(evt.Message.ChannelID == h.checkoutChannel.ID && !h.hasJudgeRole(evt.Message.Member)) {
		var message string

		if h.hasJudgeRole(evt.Message.Member) {
			message = fmt.Sprintf(
				"Check-in and check-out can only happen in the %v or %v channels.",
				h.checkinChannel.Mention(),
				h.checkoutChannel.Mention(),
			)
		} else {
			message = fmt.Sprintf("Check-in can only happen in the %v channel.", h.checkinChannel.Mention())
		}

		h.t.ReplyMessage(evt.Message, "you can't do that here. %v", message)
		h.t.RejectMessage(evt.Message)
		return
	}

	if !h.hasJudgeRole(evt.Message.Member) && checkout.Match(rawMessage) {
		h.t.ReplyMessage(evt.Message, "you can't do that. Only judges can check out.")
		h.t.RejectMessage(evt.Message)
		return
	}

	direction := "in"
	if checkout.Match(rawMessage) {
		direction = "out"
	}

	id, speaker, err := h.t.database.ParticipantFromDiscord(evt.Message.Author.ID.String())
	if err != nil {
		log.Printf("error finding participant: %v", err.Error())

		h.t.ReplyMessage(
			evt.Message,
			"there was an error checking you %v. Please ask for help in %v.",
			direction,
			h.techHelpChannel.Mention(),
		)
		h.t.RejectMessage(evt.Message)
		return
	}

	if checkout.Match(rawMessage) {
		err = h.t.tabbycat.CheckOutAdjudicator(id)
	} else {
		err = h.t.tabbycat.CheckIn(id, speaker)
	}

	if err != nil {
		log.Printf("error checking %v participant: %v", direction, err.Error())
		h.t.ReplyMessage(
			evt.Message,
			"there was an error checking you %v. Please ask for help in %v.",
			direction,
			h.techHelpChannel.Mention(),
		)
		h.t.RejectMessage(evt.Message)
		return
	}

	h.t.AcknowledgeMessage(evt.Message)
	if chicken.Match(rawMessage) {
		h.t.reactMessage(evt.Message, "🐓")
	}
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
		} else if channel.Name == "adjudicator-availability" {
			h.checkoutChannel = channel
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
		} else if role.Name == "Judge" {
			h.judgeRole = role
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

func (h *CheckinHandler) hasJudgeRole(member *disgord.Member) bool {
	for _, role := range member.Roles {
		if role == h.judgeRole.ID {
			return true
		}
	}

	return false
}
