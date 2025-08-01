package entity

import "time"

type Product struct {
	ID          int
	Title       string
	Price       int
	Description string
	Available   bool
	ImageURL    string
	CreatedAt   time.Time
}
