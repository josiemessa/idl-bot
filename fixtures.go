package main

import (
	"fmt"
	"log"
	"time"
)

type Fixture struct {
	Date     string `yaml:"date"`
	Teams    string `yaml:"teams"`
	Reminder time.Duration
}

// Given a list of fixtures, set a timer to figure out when to post in the channel
func (f *Fixture) calculateFixtureReminder() error {
	// Calculate 10am on the closest weekday to a fixture date
	// Layout has to be based of 2nd Jan 2006 because why not
	dateAndTime := fmt.Sprintf("%s 10", f.Date)
	t, err := time.Parse("02/01/2006 03", dateAndTime)
	if err != nil {
		return fmt.Errorf("Error parsing fixture time: %s", err)
	}
	// Probably don't need tis
	// reminderTime := adjustForWeekends(t)

	timeToReminder := time.Until(t)
	if timeToReminder < 0 {
		return fmt.Errorf("Too late to remind for fixture on %s", f.Date)
	}
	f.Reminder = timeToReminder
	return nil
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
