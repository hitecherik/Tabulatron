package pundit

import (
	"context"
	"log"
	"sync"
	"sync/atomic"

	"github.com/andersfylling/disgord"
)

const bufferSize int = 16

type Pundit struct {
	clients  []*disgord.Client
	channels []chan reaction
	counter  uint64
	wg       sync.WaitGroup
}

type reaction struct {
	channelId disgord.Snowflake
	messageId disgord.Snowflake
	emoji     string
}

func (p *Pundit) AddClient(client *disgord.Client) {
	p.clients = append(p.clients, client)
	p.channels = append(p.channels, make(chan reaction, bufferSize))

	p.wg.Add(1)
	go p.listen(len(p.clients) - 1)
}

func (p *Pundit) listen(client int) {
	for r := range p.channels[client] {
		err := p.clients[client].CreateReaction(context.Background(), r.channelId, r.messageId, r.emoji)
		if err != nil {
			log.Printf("Error sending reaction: %+v\n", r)
		}
	}

	p.wg.Done()
}

func (p *Pundit) SendReaction(channelId, messageId disgord.Snowflake, emoji string) {
	channel := atomic.AddUint64(&p.counter, 1) % uint64(len(p.channels))

	p.channels[channel] <- reaction{channelId, messageId, emoji}
}

func (p *Pundit) Wait() {
	for _, channel := range p.channels {
		close(channel)
	}

	p.wg.Wait()
}
