package tabulatron

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/andersfylling/disgord"
)

const (
	prepMinutes int           = 15
	minute      time.Duration = time.Second * 60
)

var (
	infoslide *regexp.Regexp = regexp.MustCompile(`^!infoslide\s*\d+$`)
	motion    *regexp.Regexp = regexp.MustCompile(`^!motion\s*\d+$`)
	roundId   *regexp.Regexp = regexp.MustCompile(`\s*(\d+)$`)
)

type MotionHandler struct {
	t              *Tabulatron
	tabRole        *disgord.Role
	motionsChannel *disgord.Channel
}

func NewMotionHandler(t *Tabulatron) *MotionHandler {
	return &MotionHandler{
		t: t,
	}
}

func (h *MotionHandler) CanHandle(_ disgord.Session, evt *disgord.MessageCreate) bool {
	message := []byte(evt.Message.Content)

	return infoslide.Match(message) || motion.Match(message)
}

func (h *MotionHandler) Handle(s disgord.Session, evt *disgord.MessageCreate) {
	rawMessage := []byte(evt.Message.Content)

	if err := h.populateRoles(evt); err != nil {
		log.Printf("error populating roles: %v", err.Error())
		return
	}

	if !h.hasTabRole(evt.Message.Member) {
		h.t.ReplyMessage(evt.Message, "you can't ask me to do that.")
		h.t.RejectMessage(evt.Message)
		return
	}

	if err := h.populateChannels(evt); err != nil {
		log.Printf("error populating channels: %v", err.Error())
		return
	}

	matches := roundId.FindSubmatch(rawMessage)
	id, err := strconv.ParseUint(string(matches[1]), 10, 64)
	if err != nil {
		log.Printf("error extracting round: %v", err.Error())
		h.t.ReplyMessage(evt.Message, "there was an error parsing your request.")
		h.t.RejectMessage(evt.Message)
		return
	}

	round, err := h.t.tabbycat.GetRound(id)
	if err != nil {
		log.Printf("error fetching round: %v", err.Error())
		h.t.ReplyMessage(evt.Message, "I couldn't find any information about that round.")
		h.t.RejectMessage(evt.Message)
		return
	}

	message := ""
	roundName := round.Name

	if !strings.HasPrefix(roundName, "Round ") {
		roundName = fmt.Sprintf("The %v", roundName)
	}

	if infoslide.Match(rawMessage) {
		if round.Motion.InfoSlide == "" {
			message = fmt.Sprintf("There is no info slide for %v.", roundName)
		} else {
			message = fmt.Sprintf("@everyone\nThe info slide for **%v** is:\n\n%v", roundName, round.Motion.InfoSlide)
		}
	} else {
		if round.Motion.Motion == "" {
			message = fmt.Sprintf("There is no motion for %v.", roundName)
		} else {
			message = fmt.Sprintf("@everyone\nThe motion for **%v** is:\n\n%v", roundName, round.Motion.Motion)
		}
	}

	if _, err := h.t.discord.SendMsg(context.Background(), h.motionsChannel.ID, message); err != nil {
		log.Printf("error announcing round: %v", err.Error())
	}

	err = h.t.discord.DeleteMessage(context.Background(), evt.Message.ChannelID, evt.Message.ID)
	if err != nil {
		log.Printf("error deleting message: %v", err.Error())
	}

	if motion.Match(rawMessage) {
		err := h.t.tabbycat.ReleaseMotion(id, time.Now().Add(time.Duration(prepMinutes)*minute))
		if err != nil {
			log.Printf("error releasing motion on tabbycat: %v", err.Error())
		}

		go func() {
			timeLeft := prepMinutes

			msg, err := h.t.discord.SendMsg(context.Background(), h.motionsChannel.ID, generatePrepTimeMessage(timeLeft))
			if err != nil {
				log.Printf("timer broke: %v", err.Error())
				return
			}

			ticker := time.NewTicker(minute)

			for {
				<-ticker.C
				timeLeft -= 1

				if timeLeft == 0 {
					ticker.Stop()
					err = h.t.discord.DeleteMessage(context.Background(), msg.ChannelID, msg.ID)
					if err != nil {
						log.Printf("error deleting message: %v", err.Error())
					}

					_, err = h.t.discord.SendMsg(
						context.Background(),
						h.motionsChannel.ID,
						fmt.Sprintf("@everyone Prep time for %v over!", roundName),
					)
					if err != nil {
						log.Printf("error sending message: %v", err.Error())
					}

					return
				} else {
					msg, err = h.t.discord.UpdateMessage(context.Background(), msg.ChannelID, msg.ID).
						SetContent(generatePrepTimeMessage(timeLeft)).
						Execute()
					if err != nil {
						ticker.Stop()
						log.Printf("error updating message: %v", err.Error())
						return
					}
				}
			}
		}()
	}
}

func (h *MotionHandler) populateChannels(evt *disgord.MessageCreate) error {
	if h.motionsChannel != nil {
		return nil
	}

	channels, err := h.t.discord.GetGuildChannels(context.Background(), evt.Message.GuildID)
	if err != nil {
		return err
	}

	for _, channel := range channels {
		if channel.Name == "motions-and-draw" {
			h.motionsChannel = channel
			break
		}
	}

	return nil
}

func (h *MotionHandler) populateRoles(evt *disgord.MessageCreate) error {
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

func (h *MotionHandler) hasTabRole(member *disgord.Member) bool {
	for _, role := range member.Roles {
		if role == h.tabRole.ID {
			return true
		}
	}

	return false
}

func generatePrepTimeMessage(timeLeft int) string {
	verb := "are"
	noun := "minutes"

	if timeLeft == 1 {
		verb = "is"
		noun = "minute"
	}

	return fmt.Sprintf("There %v **%v %v** of prep time left.", verb, timeLeft, noun)
}
