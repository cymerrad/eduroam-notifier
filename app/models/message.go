package models

import (
	"fmt"
	"time"

	"github.com/revel/revel"
)

type Message struct {
	ID        string
	Message   string
	Timestamp time.Time
	EventMessageFields
}

func (u *Message) String() string {
	return fmt.Sprintf("Message(%d, %s)", u.ID, u.Message)
}

func (u *Message) Validate(v *revel.Validation) {
	v.Required(u.Message)
}
