package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Generate fixtures dates for the next 6 days
// so we can guarantee we test the weekend stuff
func generateTestFixtures() []*Fixture {
	now := time.Now()

	fixtures := []*Fixture{}
	teams := []string{"a v b", "b v c", "c v a"}
	for i := 1; i < 7; i++ {
		next := now.AddDate(0, 0, i)
		y, m, d := next.Date()
		f := &Fixture{
			Date:  fmt.Sprintf("%.2d/%.2d/%.4d", d, m, y),
			Teams: teams[i%3],
		}
		fixtures = append(fixtures, f)
	}

	return fixtures
}

// return the indices that saturday and sunday
// will fall on so that we can account for that in tests
func weekendIndices() (sat int, sun int) {
	switch time.Now().Weekday() {
	case time.Monday:
		return 4, 5
	case time.Tuesday:
		return 3, 4
	case time.Wednesday:
		return 2, 3
	case time.Thursday:
		return 1, 2
	case time.Friday:
		return 0, 1
	case time.Saturday:
		return 0, 6
	case time.Sunday:
		return 5, 6
	}
	return 0, 0
}

func TestCalculateFixtureReminders(t *testing.T) {
	fixtures := generateTestFixtures()
	sat, sun := weekendIndices()

	for i, f := range fixtures {
		err := f.calculateFixtureReminder()
		require.Nil(t, err, "Unexpected error calculating reminder time")

		// Calculate the expected length of the reminder
		expected, _ := time.Parse("02/01/2006 03", fmt.Sprintf("%s 10", f.Date))
		// account for weekends, work out when the friday is
		if i == sat {
			expected = expected.AddDate(0, 0, -1)
		} else if i == sun {
			expected = expected.AddDate(0, 0, -2)
		}
		expectedDuration := time.Until(expected)

		// In case you're running this on a toaster, allow a 5 second delta
		require.InDelta(t, expectedDuration.Seconds(), f.Reminder.Seconds(), 5, "Unexpected reminder duration for fixture on %s. Expected %s, got %s", f.Date, expectedDuration, f.Reminder)
	}
}
