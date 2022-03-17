package types

import "time"

// DateTime is RFC3339 time format
type DateTime string

func (t DateTime) String() string {
	return string(t)
}

// IsZero returns true if DateTime has zero value
func (t DateTime) IsZero() bool {
	return t == ""
}

// ToTime returns the time value that this DateTime represents
func (t DateTime) ToTime() (time.Time, error) {
	return time.Parse(time.RFC3339, t.String())
}
