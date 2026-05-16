package domain

import (
	"errors"
	"time"
)

var (
	ErrEventNotFound = errors.New("event not found")
	ErrInvalidEvent  = errors.New("invalid event")
)

type Event struct {
	ID           string
	Sport        string
	Category     string
	Competition  string
	Title        string
	Description  string
	StartTime    time.Time
	EndTime      time.Time
	Status       string
	Country      string
	City         string
	Venue        string
	Participants []string
	Tags         []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type EventFilter struct {
	Sport       string
	Category    string
	Competition string
	Status      string
	Country     string
	City        string
	Tag         string
	Limit       int
	Offset      int
}
