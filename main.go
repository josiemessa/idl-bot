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

type Bot interface {
	Connect(string, string) error
	Run()
	SendMessage(string, bool)
}

var bot Bot

const RATE_LIMIT = 1 * time.Second

func main() {
	// Grab command line parameters for start up and verify
	discordKey := flag.String("discordkey", "REQUIRED", "Discord integration key")
	idlServer := flag.String("server", "idl", "Discord server/guild name for IDL ('IDL' by default)")
	fixturesFile := flag.String("fixtures", "fixtures.json", "Location of fixtures JSON ('fixtures.json by default)")
	flag.Parse()
	flag.VisitAll(func(f *flag.Flag) {
		if f.Value.String() == "REQUIRED" {
			flag.PrintDefaults()
			log.Fatalf(f.Name, "is a required parameter. See usage")
		}
	})

	log.SetFlags(log.LstdFlags)

	fixtures, err := openFixturesFile(*fixturesFile)
	if err != nil {
		log.Println(err)
	}

	bot = &DiscordBot{}
	if err := bot.Connect(*discordKey, *idlServer); err != nil {
		log.Fatal(err)
	}

	// Set the reminders for the fixtures
	SetReminders(fixtures)
	// Start the loop to listen for incoming events
	bot.Run()
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
func SetReminders(fixtures []*Fixture) {
	for _, f := range fixtures {
		err := f.calculateFixtureReminder()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("Setting reminder for %s for fixture %s %s\n", f.Reminder, f.Teams, f.Date)
		// need to take a copy of f here as when the timer expires, f will be
		// pointing to the last entry in the slice of fixtures
		fixture := *f
		go time.AfterFunc(fixture.Reminder, func() {
			FixtureAlert(&fixture)
		})
	}
}

func FixtureAlert(f *Fixture) {
	t, err := time.Parse("02/01/2006", f.Date)
	if err != nil {
		log.Println(err)
		return
	}
	message := fmt.Sprintf("*FIXTURE REMINDER*: %s playing at 8pm %s %s", f.Teams, t.Weekday(), f.Date)
	bot.SendMessage(message, true)
}
