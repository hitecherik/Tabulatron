package hermes

import (
	"context"
	"log"

	"github.com/andersfylling/disgord"
)

type Hermes struct {
	client   *disgord.Client
	queue    chan message
	finished chan bool
}

type message struct {
	to      disgord.Snowflake
	content string
}

func New(client *disgord.Client) *Hermes {
	return &Hermes{client, make(chan message), make(chan bool, 1)}
}

func (h *Hermes) Listen() {
	for message := range h.queue {
		channel, err := h.client.CreateDM(context.Background(), message.to)
		if err != nil {
			log.Printf("error creating DM: %v", err.Error())
			continue
		}

		_, err = channel.SendMsgString(context.Background(), h.client, message.content)
		if err != nil {
			log.Printf("error sending message: %v", err.Error())
			continue
		}
	}

	h.finished <- true
}

func (h *Hermes) SendMessage(to disgord.Snowflake, content string) {
	h.queue <- message{to, content}
}

func (h *Hermes) Wait() {
	close(h.queue)
	<-h.finished
}
