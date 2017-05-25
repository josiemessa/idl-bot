package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/nlopes/slack"
)

// Bot extends the RTM API so we can just call RTM functions on it
type Bot struct {
	rtm       *slack.RTM
	channel   string // Channel ID for the IDL channel
	Reminders []time.Timer
}

var bot Bot

func main() {
	// Grab command line parameters for start up and verify
	slackKey := flag.String("key", "REQUIRED", "Slack integration key")
	idlChannelName := flag.String("channel", "idl", "Channel name for IDL channel ('idl' by default)")
	fixturesFile := flag.String("fixtures", "fixtures.json", "Location of fixtures JSON")
	flag.Parse()
	flag.VisitAll(func(f *flag.Flag) {
		if f.Value.String() == "REQUIRED" {
			flag.PrintDefaults()
			log.Fatalf(f.Name, "is a required parameter. See usage")
		}
	})

	fixtures, err := openFixturesFile(*fixturesFile)
	if err != nil {
		log.Println(err)
	}

	// Create new RTM connection
	// set UseRTMStart to false otherwise to force the API to use
	// rtm.connect, which doesn't pull as much info as start
	log.Println("Logging into slack")
	api := slack.New(*slackKey)
	bot.rtm = api.NewRTMWithOptions(&slack.RTMOptions{UseRTMStart: false})
	go bot.rtm.ManageConnection()

	// Get the channel ID for the IDL channel so we can post messages in there
	channels, err := api.GetChannels(false)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	for _, channel := range channels {
		if channel.Name == *idlChannelName {
			// Bot.channel
		}
	}

	// Set the reminders for the fixtures
	bot.SetReminders(fixtures)
	// Start the loop to listen for incoming events
	bot.run()
}

func openFixturesFile(fixturesFile string) ([]*Fixture, error) {
	fixtures := []*Fixture{}

	if _, err := os.Stat(fixturesFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("Fixtures file does not exist, skipping")
	}
	dat, err := ioutil.ReadFile(fixturesFile)
	if err != nil {
		return nil, fmt.Errorf("Skipping fixtures as error opening fixtures file '%s': %s", fixturesFile, err)
	}
	err = json.Unmarshal(dat, &fixtures)
	if err != nil {
		return nil, fmt.Errorf("Skipping fixtures as error unmarshalling fixtures file '%s': %s", fixturesFile, err)
	}
	return fixtures, nil

}

func (b *Bot) run() {

	for msg := range bot.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			log.Println("Hello event")
			// bot.rtm.SendMessage(bot.rtm.NewOutgoingMessage("A quest? A quest! A-questing I shall go!", bot.rtm.channel))

		case *slack.ConnectedEvent:
			// TODO: add hello message when conencted into #idl only
			log.Println("Connected event")

		case *slack.MessageEvent:
			// fmt.Printf("Message: %v\n", ev)
			log.Println("Message event")

		case *slack.RTMError:
			log.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			log.Printf("Invalid credentials")
			return

		default:

			// Ignore other events..
			// fmt.Printf("Unexpected: %v\n", msg.Data)
		}
	}
}

func (b *Bot) SetReminders(fixtures []*Fixture) {
	for _, f := range fixtures {
		err := f.calculateFixtureReminder()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Setting reminder for %s for fixture %s (%s)\n", f.Reminder, f.Teams, f.Date)
		go time.AfterFunc(f.Reminder, func() {
			b.FixtureAlert(f)
		})
	}
}

func (b *Bot) FixtureAlert(f *Fixture) {
	message := fmt.Sprintf("*FIXTURE REMINDER*: %s playing at 8pm %s", f.Teams, f.Date)
	bot.rtm.SendMessage(bot.rtm.NewOutgoingMessage(message, bot.channel))
}
