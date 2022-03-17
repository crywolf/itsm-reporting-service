package mocks

import (
	"time"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/types"
)

// FixedClock is used in testing. It implements Clock interface used in mock storages to create timestamp.
type FixedClock struct {
	currentTime time.Time
}

// NewFixedClock return new clock with preset default time
func NewFixedClock() *FixedClock {
	tz, err := time.LoadLocation("Europe/Prague")
	if err != nil {
		panic(err)
	}

	return &FixedClock{
		currentTime: time.Date(2021, 4, 1, 12, 34, 56, 78, tz),
	}
}

// SetTime sets current time of the clock
func (c *FixedClock) SetTime(t time.Time) {
	c.currentTime = t
}

// AddTime adds duration to the current time of the clock
func (c *FixedClock) AddTime(d time.Duration) {
	c.currentTime = c.currentTime.Add(d)
}

// Now returns fixed time
func (c *FixedClock) Now() time.Time {
	return c.currentTime
}

// NowFormatted returns time in RFC3339 format
func (c *FixedClock) NowFormatted() types.DateTime {
	return types.DateTime(c.Now().Format(time.RFC3339))
}
