package tabulatron

import (
	"context"
	"log"
	"regexp"

	"github.com/andersfylling/disgord"
	"github.com/hitecherik/Tabulatron/internal/util"
)

const (
	clearRaw string = `^!clear\s*(\d{6})$`
)

var (
	clear *regexp.Regexp
)

type ClearHandler struct {
	t       *Tabulatron
	tabRole *disgord.Role
}

func init() {
	var err error

	clear, err = regexp.Compile(clearRaw)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func NewClearHandler(t *Tabulatron) *ClearHandler {
	return &ClearHandler{
		t: t,
	}
}

func (h *ClearHandler) CanHandle(_ disgord.Session, evt *disgord.MessageCreate) bool {
	message := []byte(evt.Message.Content)

	return clear.Match(message)
}

func (h *ClearHandler) Handle(s disgord.Session, evt *disgord.MessageCreate) {
	rawMessage := []byte(evt.Message.Content)

	if err := h.populateRoles(evt); err != nil {
		log.Printf("error populating roles: %v", err.Error())
		return
	}

	if !h.hasTabRole(evt.Message.Member) {
		h.t.ReplyMessage(evt.Message, "you can't ask me to do that.")
		h.t.RejectMessage(s, evt.Message)
		return
	}

	matches := clear.FindSubmatch(rawMessage)
	code := string(matches[1])

	discord, err := h.t.database.ClearParticipantFromBarcode(code)
	if err != nil {
		log.Printf("error clearing participant: %v", err.Error())
		h.t.ReplyMessage(evt.Message, "there was an error doing that.")
		h.t.RejectMessage(s, evt.Message)
		return
	}

	snowflake, err := util.StringToSnowflake(discord)
	if err != nil {
		log.Printf("error converting to snowflake: %v", err.Error())
		h.t.ReplyMessage(evt.Message, "there was an error resetting the user.")
		h.t.RejectMessage(s, evt.Message)
		return
	}

	err = h.t.discord.
		UpdateGuildMember(context.Background(), evt.Message.GuildID, snowflake).
		DeleteNick().
		SetRoles([]disgord.Snowflake{}).
		Execute()
	if err != nil {
		log.Printf("error resetting user: %v", err.Error())
		h.t.ReplyMessage(evt.Message, "there was an error resetting the user.")
		h.t.RejectMessage(s, evt.Message)
		return
	}

	h.t.AcknowledgeMessage(s, evt.Message)
}

func (h *ClearHandler) populateRoles(evt *disgord.MessageCreate) error {
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

func (h *ClearHandler) hasTabRole(member *disgord.Member) bool {
	for _, role := range member.Roles {
		if role == h.tabRole.ID {
			return true
		}
	}

	return false
}
