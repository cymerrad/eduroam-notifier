package models

import (
	"time"
)

type OptOut struct {
	ID       int
	Created  time.Time
	Mac      string
	Pesel    string
	Username string
	Action   string
	Comment  string
}
