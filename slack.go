package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nlopes/slack"
)

// Bot extends the RTM API so we can just call RTM functions on it
type Bot struct {
	s         *SlackBot
	Reminders []time.Timer
}

type SlackBot struct {
	client  *slack.Client
	rtm     *slack.RTM
	channel string // Channel ID for the IDL channel
}

var bot Bot

func startSlacking(key string, channelName string) {
	// Create new RTM connection
	// set UseRTMStart to false otherwise to force the API to use
	// rtm.connect, which doesn't pull as much info as start
	log.Println("Logging into slack")
	client := slack.New(key)
	rtm := client.NewRTMWithOptions(&slack.RTMOptions{UseRTMStart: false})
	s := &SlackBot{
		rtm:    rtm,
		client: client,
	}
	bot.s = s
	go s.rtm.ManageConnection()
	s.findDotaChannelID(channelName)
}

func (s *SlackBot) findDotaChannelID(channelName string) {
	// Get the channel ID for the IDL channel so we can post messages in there
	channels, err := s.client.GetChannels(false)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	for _, channel := range channels {
		if channel.Name == channelName {
			s.channel = channel.ID
		}
	}
}

func (s *SlackBot) Run() {
	for msg := range s.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			log.Println("Hello event")
			// s.rtm.SendMessage(s.rtm.NewOutgoingMessage("A quest? A quest! A-questing I shall go!", s.channel))

		case *slack.ConnectedEvent:
			log.Println("Connected event")

		case *slack.MessageEvent:
			// TODO: Add some commands in here
			log.Println("Message event")
			if strings.Contains(ev.Msg.Text, "sing") {
				s.rtm.SendMessage(s.rtm.NewOutgoingMessage("Somebody once told me...", s.channel))

			}

		case *slack.RTMError:
			log.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			log.Printf("Invalid credentials")
			return

		default:
			// Ignore other events..
		}
	}
}

func (s *SlackBot) SendMessage(message string, notify bool) {
	if notify {
		message = fmt.Sprintf("<!here|here> %s", message)
	}
	s.rtm.SendMessage(s.rtm.NewOutgoingMessage(message, s.channel))
}
