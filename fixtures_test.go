package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAdjustForWeekends(t *testing.T) {
	layout := "Monday 02/01/2006 03"
	notAdjusted := []string{
		"Friday 09/06/2017 10",
		"Monday 12/06/2017 10",
	}
	adjusted := []string{
		"Saturday 10/06/2017 10",
		"Sunday 11/06/2017 10",
	}

	for _, in := range notAdjusted {
		p, _ := time.Parse(layout, in)
		actual := adjustForWeekends(p)
		// Don't expect time to have been adjusted for weekdays,
		// so just expected is the same as the input to adjustForWeekends
		require.EqualValues(t, p, actual, "Unepected adjusted time from adjustForWeekends")
	}

	expected, _ := time.Parse(layout, "Friday 09/06/2017 10")
	for _, in := range adjusted {
		p, _ := time.Parse(layout, in)
		actual := adjustForWeekends(p)
		// Expect weekends to be adjusted to Friday
		require.EqualValues(t, expected, actual, "Incorrect adjusted time from adjustForWeekends")
	}
}
