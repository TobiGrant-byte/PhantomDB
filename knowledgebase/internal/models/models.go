package models

import "time"

type Note struct {
	ID        string
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Bookmark struct {
	ID          string
	URL         string
	Title       string
	Description string
	CreatedAt   time.Time
}