package entity

import "time"

type Event struct {
	ID              int
	UserID          int
	Title           string
	Description     string
	Place           string
	Participants    int
	MaxParticipants int
	Date            time.Time
	Active          bool
	Images          []Image
}
