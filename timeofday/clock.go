package timeofday

import (
	"database/sql/driver"
	"time"
)

const (
	// DateFormat displays how the Clock renders
	DateFormat = "15:04:05"
)

// Clock represents a date and timezone independent concept of hour, minute, and second
type Clock int

// Today returns the Clock for the current day
func (c Clock) Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), c.Hour(), c.Minute(), c.Second(), 0, now.Location())
}

// Second returns the clock second
func (c Clock) Second() int {
	return int(c % 60)
}

// Minute returns the clock minute
func (c Clock) Minute() int {
	return int((c % 3600) / 60)
}

// Hour returns the clock hour
func (c Clock) Hour() int {
	return int(c / 3600)
}

// String renders the Clock as a string
func (c Clock) String() string {
	return c.Today().Format(DateFormat)
}

// GTE returns true if clock >= t
func (c Clock) GTE(t time.Time) bool {
	return c >= Time(t)
}

// GT returns true if clock > t
func (c Clock) GT(t time.Time) bool {
	return c > Time(t)
}

// LTE returns true if clock <= t
func (c Clock) LTE(t time.Time) bool {
	return c <= Time(t)
}

// LT returns true if clock > t
func (c Clock) LT(t time.Time) bool {
	return c < Time(t)
}

// EQ returns true if clock == t
func (c Clock) EQ(t time.Time) bool {
	return c == Time(t)
}

// Parse instantiates a Clock from a string representation
func Parse(content string) (Clock, error) {
	t, err := time.Parse(DateFormat, content)
	if err != nil {
		return 0, err
	}

	return Time(t), nil
}

// Time converts the time into a Clock
func Time(t time.Time) Clock {
	return Clock(t.Hour()*60*60 + t.Minute()*60 + t.Second())
}

// Value persistes Clock to a database as a string
func (c Clock) Value() (driver.Value, error) {
	return c.String(), nil
}

// Scan reads a Clock from the db
func (c *Clock) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	switch v := src.(type) {
	case []byte:
		clock, err := Parse(string(v))
		if err != nil {
			return err
		}
		*c = clock

	case string:
		clock, err := Parse(v)
		if err != nil {
			return err
		}
		*c = clock
	}

	return nil
}
