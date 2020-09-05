package tabulatron

import (
	"context"
	"fmt"
	"log"

	"github.com/andersfylling/disgord"
	"github.com/hitecherik/Imperial-Online-IV/internal/db"
	"github.com/hitecherik/Imperial-Online-IV/pkg/tabbycat"
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
}

func New(discord *disgord.Client, database *db.Database, tabbycat *tabbycat.Tabbycat) *Tabulatron {
	t := &Tabulatron{discord, database, tabbycat, []MessageHandler{}}
	t.handlers = append(t.handlers, NewRegHandler(t), NewCheckinHandler(t))

	return t
}

func (t *Tabulatron) HandleMessage(s disgord.Session, evt *disgord.MessageCreate) {
	for _, handler := range t.handlers {
		if handler.CanHandle(s, evt) {
			handler.Handle(s, evt)
			return
		}
	}

	log.Printf("could not find handler for message '%v' from '%v'", evt.Message.Content, evt.Message.Member.Nick)
}

func (t *Tabulatron) ReplyMessage(message *disgord.Message, reply string, a ...interface{}) *disgord.Message {
	fullReply := fmt.Sprintf(reply, a...)
	m, err := message.Reply(context.Background(), t.discord, fmt.Sprintf("%v, %v", message.Author.Mention(), fullReply))

	if err != nil {
		log.Printf("Error sending message '%v': %v", reply, err.Error())
	}

	return m
}

func (t *Tabulatron) AcknowledgeMessage(s disgord.Session, message *disgord.Message) {
	if err := message.React(context.Background(), s, "✅"); err != nil {
		log.Printf("error reacting: %v\n", err.Error())
	}
}
