package util

import (
	"fmt"
	"time"
)

// Calculates the amount time until a certain timestamp.
// Returns a string containing the number of days and hours left.
// If days or hours is 0, that part is omitted.
func TimeLeft(deadline time.Time) string {
	d := time.Until(deadline)

	// Deadline has passed
	if d <= 0 {
		return "Expired"
	}

	// Calc num days and num hours left
	total := int(d.Hours())
	days, hours := total/24, total%24

	switch {
	case days > 0 && hours > 0:
		return fmt.Sprintf("%d day(s) and %d hour(s) left", days, hours)
	case days > 0 && hours <= 0:
		return fmt.Sprintf("%d day(s) left", days)
	case days <= 0 && hours > 0:
		return fmt.Sprintf("%d hour(s) left", hours)
	default:
		return "Less than 1 hour left"
	}
}
