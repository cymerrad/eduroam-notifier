package models

import (
	"fmt"
	"time"

	"github.com/revel/revel"
)

type Message struct {
	ID        int // non-PK?
	EventID   int
	Message   string
	Timestamp time.Time
	Mac       string
	Pesel     string
	Username  string
	Action    string
	Realm     string
	Facility  string
}

func (u *Message) String() string {
	return fmt.Sprintf("Message(%d, %s)", u.ID, u.Message)
}

func (u *Message) Validate(v *revel.Validation) {
	v.Required(u.Message)

	v.MacAddr(u.Mac)
}
