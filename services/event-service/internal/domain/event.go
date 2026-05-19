package domain

import (
	"errors"
	"time"
)

var (
	ErrEventNotFound  = errors.New("event not found")
	ErrInvalidEvent   = errors.New("invalid event")
	ErrEventForbidden = errors.New("event forbidden")
	ErrEventFull      = errors.New("event full")
)

type Event struct {
	ID                string
	CreatorID         string
	Sport             string
	Category          string
	Competition       string
	Title             string
	Description       string
	Location          string
	StartTime         time.Time
	EndTime           time.Time
	Status            string
	Country           string
	City              string
	MaxParticipants   int32
	ParticipantsCount int32
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type EventFilter struct {
	Sport         string
	Category      string
	Competition   string
	Status        string
	StartTimeFrom time.Time
	StartTimeTo   time.Time
	Country       string
	City          string
	Limit         int
	Offset        int
}
