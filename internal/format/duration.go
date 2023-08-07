package format

import (
	"time"
)

func DurationAsAdj(d time.Duration) string {
	switch {
	case d > oneYear:
		return "yearly"
	case d > oneMonth:
		return "monthly"
	case d > 2*oneWeek:
		return "biweekly"
	case d > oneWeek:
		return "weekly"
	case d > oneDay:
		return "daily"
	case d > time.Hour:
		return "hourly"
	}

	return ""
}
