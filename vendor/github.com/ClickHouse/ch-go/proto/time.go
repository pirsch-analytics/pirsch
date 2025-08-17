package proto

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Time32 represents time of day in seconds since midnight.
type Time32 int32

// Time64 represents time of day with precision as a decimal with configurable scale.
type Time64 int64

func (t Time32) String() string {
	seconds := int32(t)
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}

func (t Time64) String() string {
	// Time64 stores as decimal with scale, convert to nanoseconds for display
	scale := int64(1e9) // Default to nanosecond scale for display
	value := int64(t)

	// Convert to nanoseconds since midnight
	seconds := value / scale
	nanos := (value % scale) * (1e9 / scale)

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	return fmt.Sprintf("%02d:%02d:%02d.%09d", hours, minutes, secs, nanos)
}

func ParseTime32(s string) (Time32, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format: %s", s)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, err
	}

	totalSeconds := int32(hours*3600 + minutes*60 + seconds)
	return Time32(totalSeconds), nil
}

func ParseTime64(s string) (Time64, error) {
	// Parse time string like "12:34:56.789"
	timePart, fractionalStr, ok := strings.Cut(s, ".")
	if !ok {
		fractionalStr = ""
	}

	parts := strings.Split(timePart, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format: %s", s)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, err
	}

	// Calculate total seconds since midnight
	totalSeconds := int64(hours*3600 + minutes*60 + seconds)

	// Parse fractional part (default to nanoseconds scale)
	var fractional int64
	if fractionalStr != "" {
		// Pad or truncate to 9 digits (nanoseconds)
		for len(fractionalStr) < 9 {
			fractionalStr += "0"
		}
		if len(fractionalStr) > 9 {
			fractionalStr = fractionalStr[:9]
		}
		fractional, err = strconv.ParseInt(fractionalStr, 10, 64)
		if err != nil {
			return 0, err
		}
	}

	// Store as decimal with nanosecond scale
	return Time64(totalSeconds*1e9 + fractional), nil
}

func FromTime32(t time.Time) Time32 {
	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()
	totalSeconds := int32(hour*3600 + minute*60 + second)
	return Time32(totalSeconds)
}

// FromTime64 converts time.Time to Time64 with default precision (9 - nanoseconds)
func FromTime64(t time.Time) Time64 {
	return FromTime64WithPrecision(t, 9)
}

// FromTime64WithPrecision converts time.Time to Time64 with specified precision
// Time64 stores time as a decimal with configurable scale, similar to DateTime64
func FromTime64WithPrecision(t time.Time, precision int) Time64 {
	hour := t.Hour()
	minute := t.Minute()
	second := t.Second()
	nanosecond := t.Nanosecond()

	// Calculate seconds since midnight
	totalSeconds := int64(hour*3600 + minute*60 + second)

	// Calculate scale based on precision (same as DateTime64)
	scale := int64(1)
	for i := 0; i < precision; i++ {
		scale *= 10
	}

	// Convert to the appropriate scale
	// Store as: seconds * scale + fractional_part
	fractional := int64(nanosecond) * scale / 1e9
	return Time64(totalSeconds*scale + fractional)
}

func (t Time32) ToTime32() time.Time {
	seconds := int32(t)
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	return time.Date(1970, 1, 1, int(hours), int(minutes), int(secs), 0, time.UTC)
}

// ToTime converts Time64 to time.Time with default precision (9 - nanoseconds)
func (t Time64) ToTime() time.Time {
	return t.ToTimeWithPrecision(9)
}

// ToTimeWithPrecision converts Time64 to time.Time with specified precision
// Time64 stores time as a decimal with configurable scale, similar to DateTime64
func (t Time64) ToTimeWithPrecision(precision int) time.Time {
	// Calculate scale based on precision (same as DateTime64)
	scale := int64(1)
	for i := 0; i < precision; i++ {
		scale *= 10
	}

	value := int64(t)

	// Extract whole and fractional parts
	whole := value / scale
	fractional := value % scale

	// Convert fractional part to nanoseconds
	nanoseconds := fractional * 1e9 / scale

	// Convert whole part to time components
	hours := whole / 3600
	minutes := (whole % 3600) / 60
	seconds := whole % 60

	return time.Date(1970, 1, 1, int(hours), int(minutes), int(seconds), int(nanoseconds), time.UTC)
}
