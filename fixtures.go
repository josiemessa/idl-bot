package main

import (
	"fmt"
	"time"
)

type Fixture struct {
	Date     string `json:"date"`
	Teams    string `json:"teams"`
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
