package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"strings"

	"gopkg.in/yaml.v2"
)

type Bot interface {
	Connect(string, string) error
	Run()
	SendMessage(string, bool)
}

var (
	bot      Bot
	teams    []*Team
	fixtures []*Fixture
)

const RATE_LIMIT = 1 * time.Second

func main() {
	// Grab command line parameters for start up and verify
	discordKey := flag.String("discordkey", "REQUIRED", "Discord integration key")
	idlServer := flag.String("server", "idl", "Discord server/guild name for IDL ('IDL' by default)")
	dataDir := flag.String("data", "files", "Directory containing data files ('files' by default)")
	flag.Parse()
	flag.VisitAll(func(f *flag.Flag) {
		if f.Value.String() == "REQUIRED" {
			flag.PrintDefaults()
			log.Fatalf(f.Name, "is a required parameter. See usage")
		}
	})

	log.SetFlags(log.LstdFlags)

	var (
		teamdata *TeamFile
		err      error
	)
	fixtures, teamdata, err = openDataDir(*dataDir)
	if err != nil {
		log.Println(err)
	}

	teams = teamdata.PopulateTeams()

	bot = &DiscordBot{}
	if err := bot.Connect(*discordKey, *idlServer); err != nil {
		log.Fatal(err)
	}

	// Set the reminders for the fixtures
	SetReminders(fixtures)
	// Start the loop to listen for incoming events
	bot.Run()
}

func openDataDir(dir string) ([]*Fixture, *TeamFile, error) {
	fixtures := []*Fixture{}
	tf := TeamFile{}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("Data directory does not exist")
	}

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "fixtures.yml") {
			dat, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("Skipping fixtures as error opening fixtures file '%s': %s", path, err)
			}
			err = yaml.Unmarshal(dat, &fixtures)
			if err != nil {
				return fmt.Errorf("Skipping fixtures as error unmarshalling fixtures file '%s': %s", path, err)
			}
			log.Printf("Loaded %d fixtures", len(fixtures))
		} else if strings.HasSuffix(path, "teams.yml") {
			dat, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("Skipping teams as error opening teams file '%s': %s", path, err)
			}
			err = yaml.Unmarshal(dat, &tf)
			if err != nil {
				return fmt.Errorf("Skipping teams as error unmarshalling teams file '%s': %s", path, err)
			}
			log.Printf("Loaded %d teams and %d players", len(tf.Teams), len(tf.Players))
		}
		return nil
	})

	return fixtures, &tf, nil

}
