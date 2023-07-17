package model

import "time"

// PMMInstance represents a PMM instance.
type PMMInstance struct {
	ID             string
	URL            string
	APIKeySecretID string

	CreatedAt time.Time
	UpdatedAt time.Time
}
