package tabulatron

import (
	"context"
	"fmt"
	"log"

	"github.com/andersfylling/disgord"
	"github.com/hitecherik/Tabulatron/internal/db"
	"github.com/hitecherik/Tabulatron/internal/pundit"
	"github.com/hitecherik/Tabulatron/pkg/tabbycat"
)

type MessageHandler interface {
	CanHandle(disgord.Session, *disgord.MessageCreate) bool
	Handle(disgord.Session, *disgord.MessageCreate)
}

type Tabulatron struct {
	discord  *disgord.Client
	database *db.Database
	tabbycat *tabbycat.Tabbycat
	handlers []MessageHandler
	pundit   *pundit.Pundit
}

func New(discord *disgord.Client, database *db.Database, tabbycat *tabbycat.Tabbycat, p *pundit.Pundit) *Tabulatron {
	t := &Tabulatron{discord, database, tabbycat, []MessageHandler{}, p}
	t.handlers = append(t.handlers, NewRegHandler(t), NewCheckinHandler(t), NewClearHandler(t), NewMotionHandler(t), NewTabbycatRoundsHandler(t))

	return t
}

func (t *Tabulatron) HandleMessage(s disgord.Session, evt *disgord.MessageCreate) {
	for _, handler := range t.handlers {
		if handler.CanHandle(s, evt) {
			handler.Handle(s, evt)
			return
		}
	}

	log.Printf("could not find handler for message '%v' from '%v'", evt.Message.Content, evt.Message.Author.Username)
}

func (t *Tabulatron) HandleDeparture(s disgord.Session, evt *disgord.GuildMemberRemove) {
	if err := t.database.ClearParticipantFromDiscord(fmt.Sprint(evt.User.ID)); err != nil {
		log.Printf("could not clear user %v (snowflake %v): %v", evt.User.Username, evt.User.ID, err.Error())
	}
}

func (t *Tabulatron) ReplyMessage(message *disgord.Message, reply string, a ...interface{}) *disgord.Message {
	fullReply := fmt.Sprintf(reply, a...)
	m, err := message.Reply(context.Background(), t.discord, fmt.Sprintf("%v, %v", message.Author.Mention(), fullReply))

	if err != nil {
		log.Printf("Error sending message '%v': %v", reply, err.Error())
	}

	return m
}

func (t *Tabulatron) AcknowledgeMessage(message *disgord.Message) {
	t.reactMessage(message, "✅")
}

func (t *Tabulatron) RejectMessage(message *disgord.Message) {
	t.reactMessage(message, "❌")
}

func (t *Tabulatron) CreateDMAndSendMessage(snowflake disgord.Snowflake, message string) {
	channel, err := t.discord.CreateDM(context.Background(), snowflake)
	if err != nil {
		log.Printf("error creating DM with user %v: %v", snowflake, err.Error())
		return
	}

	_, err = channel.SendMsgString(context.Background(), t.discord, message)
	if err != nil {
		log.Printf("error sending DM to user %v: %v", snowflake, err.Error())
	}
}

func (t *Tabulatron) reactMessage(message *disgord.Message, reaction string) {
	t.pundit.SendReaction(message.ChannelID, message.ID, reaction)
}
