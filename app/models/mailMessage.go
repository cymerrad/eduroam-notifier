package models

import (
	"fmt"
	"time"

	"github.com/revel/revel"
	gorp "gopkg.in/gorp.v2"
)

type MailMessage struct {
	ID            string `json:"-"`
	EventID       int    `json:"event_id"`
	CreatedString string `json:"timestamp"`
	Recipient     string `json:"recipient"`
	Body          string `json:"body"`
	Error         string `json:"error,omitempty"`

	// transient
	Created time.Time `json:"-"`
}

func (u *MailMessage) String() string {
	return fmt.Sprintf("MailMessage(%d, %s)", u.EventID, u.Recipient)
}

func (u *MailMessage) Validate(v *revel.Validation) {
	v.Required(u.EventID)

}

func (o *MailMessage) PreInsert(_ gorp.SqlExecutor) error {
	o.CreatedString = o.Created.Format(time.RFC3339)

	return nil
}

func (o *MailMessage) PostGet(_ gorp.SqlExecutor) error {
	t, err := time.Parse(time.RFC3339, o.CreatedString)
	if err != nil {
		return err
	}
	o.Created = t
	return nil
}
