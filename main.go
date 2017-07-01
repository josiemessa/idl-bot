package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

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

	startSlacking(*slackKey, *idlChannelName)

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
	b.s.Run()
}

func (b *Bot) SetReminders(fixtures []*Fixture) {
	for _, f := range fixtures {
		err := f.calculateFixtureReminder()
		if err != nil {
			log.Println(err)
			continue
		}
		t, err := time.Parse("02/01/2006", f.Date)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Setting reminder for %s for fixture %s (%s %s)\n", f.Reminder, f.Teams, t.Weekday(), f.Date)
		go time.AfterFunc(f.Reminder, func() {
			b.FixtureAlert(f)
		})
	}
}

func (b *Bot) FixtureAlert(f *Fixture) {
	t, err := time.Parse("02/01/2006", f.Date)
	if err != nil {
		log.Println(err)
	}
	message := fmt.Sprintf("*FIXTURE REMINDER*: %s playing at 8pm %s %s KUNDALINI", f.Teams, t.Weekday(), f.Date)
	bot.SendMessage(message, true)

}

func (b *Bot) SendMessage(message string, notify bool) {
	b.s.SendMessage(message, notify)
}
