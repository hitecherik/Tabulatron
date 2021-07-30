package tabulatron

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/andersfylling/disgord"
	"github.com/olekukonko/tablewriter"
)

type TabbycatRoundsHandler struct {
	t       *Tabulatron
	tabRole *disgord.Role
}

func NewTabbycatRoundsHandler(t *Tabulatron) *TabbycatRoundsHandler {
	return &TabbycatRoundsHandler{
		t: t,
	}
}

func (h *TabbycatRoundsHandler) CanHandle(_ disgord.Session, evt *disgord.MessageCreate) bool {
	return evt.Message.Content == "!tabbycatrounds"
}

func (h *TabbycatRoundsHandler) Handle(s disgord.Session, evt *disgord.MessageCreate) {
	if err := h.populateRoles(evt); err != nil {
		log.Printf("error populating roles: %v", err.Error())
		return
	}

	if !h.hasTabRole(evt.Message.Member) {
		h.t.RejectMessage(evt.Message)
		h.t.ReplyMessage(evt.Message, "you can't ask me to do that.")
		return
	}

	rounds, err := h.t.tabbycat.GetRounds()
	if err != nil {
		h.t.RejectMessage(evt.Message)
		h.t.ReplyMessage(evt.Message, "there was an error fetching rounds for this tournament.")
		log.Printf("error fetching rounds: %v", err.Error())
		return
	}

	writer := &strings.Builder{}
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"ID", "Name"})

	for _, round := range rounds {
		table.Append([]string{round.Id, round.Name})
	}

	table.Render()

	_, err = h.t.discord.SendMsg(context.Background(), evt.Message.ChannelID, fmt.Sprintf("The rounds for this tournament:\n```%v```", writer.String()))
	if err != nil {
		log.Printf("error sending rounds: %v", err.Error())
	}
}

func (h *TabbycatRoundsHandler) populateRoles(evt *disgord.MessageCreate) error {
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

func (h *TabbycatRoundsHandler) hasTabRole(member *disgord.Member) bool {
	for _, role := range member.Roles {
		if role == h.tabRole.ID {
			return true
		}
	}

	return false
}
