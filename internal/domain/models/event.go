package models

import (
	"time"
)

// Event main model
type Event struct {
	ID        int64
	Owner     string
	Title     string
	Text      string
	StartTime time.Time
	EndTime   time.Time
}
