package models

import (
	"fmt"
	"time"

	"github.com/revel/revel"
)

type MailMessage struct {
	ID        string    `json:"-"`
	EventID   int       `json:"event_id"`
	Timestamp time.Time `json:"timestamp"`
	Recipient string    `json:"recipient"`
	Body      string    `json:"body"`
	Error     string    `json:"error,omitempty"`
}

func (u *Message) String() string {
	return fmt.Sprintf("Message(%d, %s)", u.ID, u.Message)
}

func (u *Message) Validate(v *revel.Validation) {
	v.Required(u.Message)

	v.MacAddr(u.Mac)
}
