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
	// Figure out if the match is on the weekend and if so, set the reminder
	// to be on a weekday
	reminderTime := t
	switch t.Weekday() {
	case time.Saturday:
		reminderTime = t.AddDate(0, 0, -1)
	case time.Sunday:
		reminderTime = t.AddDate(0, 0, -2)
	}
	timeToReminder := time.Until(reminderTime)
	if timeToReminder < 0 {
		return fmt.Errorf("Fixture has already happened %s", f.Date)
	}
	f.Reminder = timeToReminder
	return nil
}
