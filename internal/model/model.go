package model

import (
	"github.com/google/uuid"
	"time"
)

type BlogEntry struct {
	ID           uuid.UUID `json:"id"`
	CreationDate time.Time `json:"creationDate"`
	Headline     string    `json:"headline"`
	Text         string    `json:"text"`
}
