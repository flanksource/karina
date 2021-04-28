package postgres

import "time"

type ResticSnapshot struct {
	Time    time.Time
	Paths   []string
	Tags    []string
	ID      string
	ShortID string `json:"short_id"`
}
