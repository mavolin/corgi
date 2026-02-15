package html

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type DateTime string

// Month returns the [valid month string] for the given time.
//
// [valid month string]: https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#valid-month-string
func Month(t time.Time) DateTime {
	return DateTime(t.Format("2006-01"))
}

// Date returns the [valid date string] for the given time.
//
// [valid date string]: https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#valid-date-string
func Date(t time.Time) DateTime {
	return DateTime(t.Format("2006-01-02"))
}

// YearlessDate returns the [valid yearless date string] for the given time.
//
// [valid yearless date string]: https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#valid-yearless-date-string
func YearlessDate(t time.Time) DateTime {
	return DateTime(t.Format("01-02"))
}

type TimeFormat uint8

const (
	ToMinute TimeFormat = iota
	ToSecond
	ToMillisecond
)

func (f TimeFormat) toLayout(t time.Time) string {
	switch f {
	case ToMinute:
		return "15:04"
	case ToSecond:
		if t.Second() == 0 {
			return "15:04"
		}
		return "15:04:05"
	case ToMillisecond:
		ms := t.Nanosecond() / 1_000_000
		if t.Second() == 0 && ms == 0 {
			return "15:04"
		} else if ms == 0 {
			return "15:04:05"
		}
		return "15:04:05.000"
	default:
		panic(fmt.Sprintf("invalid TimeFormat %d", f))
	}
}

// Time returns the shortest possible [valid time string] for the given time.
//
// [valid time string]: https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#valid-time-string
func Time(t time.Time, f TimeFormat) DateTime {
	return DateTime(t.Format(f.toLayout(t)))
}

// LocalDateTime returns the
// [valid normalized local date and time string] for the given time.
//
// [valid normalized local date and time string]: https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#valid-normalised-local-date-and-time-string
func LocalDateTime(t time.Time, f TimeFormat) DateTime {
	return DateTime(t.Format("2006-01-02T" + f.toLayout(t)))
}

// GlobalDateTime returns the shortest possible
// [valid global date and time string] for the given time.
//
// The HTML spec does not support second-level time zone offsets, requiring
// them to be at least minute-aligned.
// For times that do have a second-level offset, the seconds will be truncated.
//
// [valid global date and time string]: https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#valid-global-date-and-time-string
func GlobalDateTime(t time.Time, f TimeFormat) DateTime {
	return DateTime(t.Format("2006-01-02T" + f.toLayout(t) + tzLayout(t)))
}

// TimeZoneOffset returns the [valid time-zone offset string] for the given
// time.
//
// The HTML spec does not support second-level time zone offsets, requiring
// them to be at least minute-aligned.
// For times that do have a second-level offset, the seconds will be truncated.
//
// [valid time-zone offset string]: https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#valid-time-zone-offset-string
func TimeZoneOffset(t time.Time) DateTime {
	return DateTime(t.Format(tzLayout(t)))
}

func tzLayout(t time.Time) string {
	_, seconds := t.Zone() // in seconds

	minutes := seconds / 60
	if minutes < 0 {
		minutes = -minutes // normalize for modulo
	}
	modMinutes := minutes % 60 // to the hour
	if modMinutes == 0 {
		return "Z07"
	}
	return "Z07:00"
}

// Week returns the [valid week string] for the given time.
//
// [valid week string]: https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#valid-week-string
func Week(t time.Time) DateTime {
	year, week := t.ISOWeek()
	return DateTime(fmt.Sprintf("%04d-W%02d", year, week))
}

// Year returns the [valid year string] for the given time.
//
// [valid year string]: https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#valid-year-string
func Year(t time.Time) DateTime {
	return DateTime(t.Format("2006"))
}

// Duration returns the [valid duration string] for the given duration.
//
// Durations below zero are not supported by the HTML specification.
// If a negative duration is given, a safe replacement is returned instead.
//
// Precision is limited to milliseconds, as the HTML specification does
// not support higher precision.
//
// [valid duration string]: https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#valid-duration-string
func Duration(d time.Duration) DateTime {
	if d < 0 {
		return "NegativeDuration"
	} else if d == 0 {
		return "PT0S"
	}

	var b strings.Builder
	b.Grow(len("P9999DT23H59M59.999S")) // realistic max size
	b.WriteByte('P')

	if days := d / (24 * time.Hour); days >= 1 {
		b.WriteString(strconv.Itoa(int(days)))
		b.WriteByte('D')
		d %= 24 * time.Hour
	}
	if d == 0 {
		return DateTime(b.String())
	}

	b.WriteByte('T')

	if hours := d / time.Hour; hours >= 1 {
		b.WriteString(strconv.Itoa(int(hours)))
		b.WriteByte('H')
		d %= time.Hour
		if d == 0 {
			return DateTime(b.String())
		}
	}

	if minutes := d / time.Minute; minutes >= 1 {
		b.WriteString(strconv.Itoa(int(minutes)))
		b.WriteByte('M')
		d %= time.Minute
		if d == 0 {
			return DateTime(b.String())
		}
	}

	seconds := d / time.Second
	d %= time.Second
	milliseconds := d / time.Millisecond
	for milliseconds > 0 && milliseconds%10 == 0 {
		milliseconds /= 10 // trim trailing zeros
	}

	switch {
	case seconds >= 1:
		b.WriteString(strconv.Itoa(int(seconds)))
		d %= time.Second
		if d == 0 {
			b.WriteByte('S')
			return DateTime(b.String())
		}
	case milliseconds >= 1:
		b.WriteByte('0')
	default:
		return DateTime(b.String())
	}
	b.WriteByte('.')
	b.WriteString(strconv.Itoa(int(milliseconds)))
	b.WriteByte('S')

	return DateTime(b.String())
}
