package models

import (
	"fmt"
	"time"

	"github.com/revel/revel"
	gorp "gopkg.in/gorp.v2"
)

type MailMessage struct {
	ID        int       `json:"-"`
	Created   time.Time `json:"-"`
	EventID   int       `json:"event_id"`
	Recipient string    `json:"recipient"`
	Body      []byte    `json:"-"`
	Error     string    `json:"error,omitempty"`

	// transient
	BodyString string `json:"body"`
}

func (u *MailMessage) String() string {
	return fmt.Sprintf("MailMessage(%d, %s)", u.EventID, u.Recipient)
}

func (u *MailMessage) Validate(v *revel.Validation) {
	v.Required(u.EventID)

}

func (o *MailMessage) PreInsert(_ gorp.SqlExecutor) error {
	o.Body = []byte(o.BodyString)

	return nil
}

func (o *MailMessage) PostGet(_ gorp.SqlExecutor) error {
	o.BodyString = string(o.Body)
	return nil
}
