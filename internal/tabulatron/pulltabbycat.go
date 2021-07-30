package tabulatron

import (
	"context"
	"fmt"
	"log"

	"github.com/andersfylling/disgord"
)

type PullTabbycatHandler struct {
	t       *Tabulatron
	tabRole *disgord.Role
}

func NewPullTabbycatHandler(t *Tabulatron) *PullTabbycatHandler {
	return &PullTabbycatHandler{
		t: t,
	}
}

func (h *PullTabbycatHandler) CanHandle(_ disgord.Session, evt *disgord.MessageCreate) bool {
	return evt.Message.Content == "!pulltabbycat"
}

func (h *PullTabbycatHandler) Handle(s disgord.Session, evt *disgord.MessageCreate) {
	if err := h.populateRoles(evt); err != nil {
		log.Printf("error populating roles: %v", err.Error())
		return
	}

	if !h.hasTabRole(evt.Message.Member) {
		h.t.RejectMessage(evt.Message)
		h.t.ReplyMessage(evt.Message, "you can't ask me to do that.")
		return
	}

	teams, err := h.t.tabbycat.GetTeams()
	if err != nil {
		h.t.RejectMessage(evt.Message)
		h.t.ReplyMessage(evt.Message, "there was an error fetching teams")
		log.Printf("error pulling teams: %v", err.Error())
	}

	logMsg, err := h.t.discord.SendMsg(context.Background(), evt.Message.ChannelID, fmt.Sprintf("Fetched %v teams", len(teams)))
	if err != nil {
		log.Printf("error sending progress message: %v", err.Error())
		return
	}

	adjudicators, err := h.t.tabbycat.GetAdjudicators()
	if err != nil {
		h.t.RejectMessage(evt.Message)
		h.t.ReplyMessage(evt.Message, "there was an error fetching adjudicators")
		log.Printf("error pulling adjudicators: %v", err.Error())
		return
	}

	logMsg, err = h.t.discord.UpdateMessage(context.Background(), logMsg.ChannelID, logMsg.ID).
		SetContent(fmt.Sprintf("Fetched %v teams\nFetched %v adjudicators", len(teams), len(adjudicators))).
		Execute()
	if err != nil {
		log.Printf("error updating progress message: %v", err.Error())
		return
	}

	if err := h.t.database.AddTeams(teams); err != nil {
		h.t.RejectMessage(evt.Message)
		h.t.ReplyMessage(evt.Message, "there was an error adding teams")
		log.Printf("error adding teams: %v", err.Error())
		return
	}

	logMsg, err = h.t.discord.UpdateMessage(context.Background(), logMsg.ChannelID, logMsg.ID).
		SetContent(fmt.Sprintf("Fetched %v teams\nFetched %v adjudicators\nInserted %v teams into database", len(teams), len(adjudicators), len(teams))).
		Execute()
	if err != nil {
		log.Printf("error updating progress message: %v", err.Error())
		return
	}

	if err := h.t.database.AddParticipants(false, adjudicators); err != nil {
		h.t.RejectMessage(evt.Message)
		h.t.ReplyMessage(evt.Message, "there was an error adding adjudicators")
		log.Printf("error adding adjudicators: %v", err.Error())
		return
	}

	_, err = h.t.discord.UpdateMessage(context.Background(), logMsg.ChannelID, logMsg.ID).
		SetContent(fmt.Sprintf("Fetched %v teams\nFetched %v adjudicators\nInserted %v teams into database\nInserted %v adjudicators into database", len(teams), len(adjudicators), len(teams), len(adjudicators))).
		Execute()
	if err != nil {
		log.Printf("error updating progress message: %v", err.Error())
		return
	}

	h.t.AcknowledgeMessage(evt.Message)
}

func (h *PullTabbycatHandler) populateRoles(evt *disgord.MessageCreate) error {
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

func (h *PullTabbycatHandler) hasTabRole(member *disgord.Member) bool {
	for _, role := range member.Roles {
		if role == h.tabRole.ID {
			return true
		}
	}

	return false
}
