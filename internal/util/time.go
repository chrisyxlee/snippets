package util

import "time"

const (
	Week = time.Hour * 24 * 7
)

const (
	TimeLayout = "2006-01-02 15:04"
	DateLayout = "Jan 2"
)

// FormatLocalTime formats the time in the local timezone with minute granularity.
func FormatLocalTime(t time.Time) string {
	return t.Local().Format(TimeLayout)
}

// ParseLocalTime parses time in the format defined by TimeLayout. The output of
// FormatLocalTime should be able to be parsed with no errors.
func ParseLocalTime(timeStr string) (time.Time, error) {
	return time.ParseInLocation(TimeLayout, timeStr, time.Local)
}

// ParseLocalTimeOrDie parses time in the format defined by TimeLayout. On error,
// this function will panic. The output of FormatLocalTime should be able to be parsed
// without dying.
func ParseLocalTimeOrDie(timeStr string) time.Time {
	t, err := ParseLocalTime(timeStr)
	Assert(err == nil, "Failed to parse time, wanted layout `%s`, got `%s`",
		TimeLayout, timeStr)
	return t
}

// FormatLocalDate formats time in the local time zone with day granularity and does not
// display the year.
func FormatLocalDate(t time.Time) string {
	return t.Local().Format(DateLayout)
}

// InTimeRange returns true if the time t is within the range [start, end).
func InTimeRange(t time.Time, start time.Time, end time.Time) bool {
	return !t.Before(start) && t.Before(end)
}
