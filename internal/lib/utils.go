package lib

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gosimple/slug"
)

const (
	Day      = 24 * time.Hour
	Week     = 7 * Day
	Month    = 30 * Day
	Year     = 12 * Month
	LongTime = 42 * Year
)

func HumanizeTime(then time.Time) string {
	return RelativeTime(then, time.Now(), "ago", "from now")
}

type RelativeTimeMagnitude struct {
	D      time.Duration
	Format string
	DivBy  time.Duration
}

var defaultMagnitudes = []RelativeTimeMagnitude{
	{time.Second, "now", time.Second},
	{2 * time.Second, "1 second %s", 1},
	{time.Minute, "%d seconds %s", time.Second},
	{2 * time.Minute, "1 minute %s", 1},
	{time.Hour, "%d minutes %s", time.Minute},
	{2 * time.Hour, "1 hour %s", 1},
	{Day, "%d hours %s", time.Hour},
	{2 * Day, "1 day %s", 1},
	{Week, "%d days %s", Day},
	{2 * Week, "1 week %s", 1},
	{Month, "%d weeks %s", Week},
	{2 * Month, "1 month %s", 1},
	{Year, "%d months %s", Month},
	{2 * Year, "1 year %s", 1},
	{LongTime, "%d years %s", Year},
}

func RelativeTime(a, b time.Time, albl, blbl string) string {
	return CustomRelativeTime(a, b, albl, blbl, defaultMagnitudes)
}

func CustomRelativeTime(a, b time.Time, albl, blbl string, magnitudes []RelativeTimeMagnitude) string {
	lbl := albl
	diff := b.Sub(a)

	if a.After(b) {
		lbl = blbl
		diff = a.Sub(b)
	}

	n := sort.Search(len(magnitudes), func(i int) bool {
		return magnitudes[i].D > diff
	})

	if n >= len(magnitudes) {
		n = len(magnitudes) - 1
	}

	mag := magnitudes[n]
	args := []interface{}{}
	escaped := false

	for _, ch := range mag.Format {
		if escaped {
			switch ch {
			case 's':
				args = append(args, lbl)
			case 'd':
				args = append(args, diff/mag.DivBy)
			}
			escaped = false
		} else {
			escaped = ch == '%'
		}
	}

	return strings.TrimSpace(fmt.Sprintf(mag.Format, args...))
}

func stripTrailingZeros(s string) string {
	offset := len(s) - 1
	for offset > 0 {
		if s[offset] == '.' {
			offset--
			break
		}
		if s[offset] != '0' {
			break
		}
		offset--
	}
	return s[:offset+1]
}

func stripTrailingDigits(s string, digits int) string {
	if i := strings.Index(s, "."); i >= 0 {
		if digits <= 0 {
			return s[:i]
		}
		i++
		if i+digits >= len(s) {
			return s
		}
		return s[:i+digits]
	}
	return s
}

// Ftoa converts a float to a string with no trailing zeros.
func Ftoa(num float64) string {
	return stripTrailingZeros(strconv.FormatFloat(num, 'f', 6, 64))
}

// FtoaWithDigits converts a float to a string but limits the resulting string
// to the given number of decimal places, and no trailing zeros.
func FtoaWithDigits(num float64, digits int) string {
	return stripTrailingZeros(stripTrailingDigits(strconv.FormatFloat(num, 'f', 6, 64), digits))
}

// CleanSlug removes non-alphanumeric characters from a string and will
// lowercase the string if setLower is true.
func CleanSlug(s string, setLower bool) string {
	slug.Lowercase = setLower
	s = slug.Make(s)

	if len(s) > 50 {
		tempCategory := strings.Split(s, "-")
		for i := range tempCategory {
			combined := strings.Join(tempCategory[:i+1], "-")
			if len(combined) > 50 {
				s = strings.Join(tempCategory[:i], "-")
			}
		}
		if len(s) > 50 {
			s = s[:50]
		}
	}

	return s
}

func Unique(x []time.Time) []time.Time {
	// Use map to record duplicates as we find them.
	encountered := map[time.Time]bool{}
	result := []time.Time{}

	for v := range x {
		if encountered[x[v]] {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[x[v]] = true
			// Append to result slice.
			result = append(result, x[v])
		}
	}

	// Return the new slice.
	return result
}
