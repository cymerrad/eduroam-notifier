package models

import (
	"fmt"
	"time"

	"github.com/revel/revel"
)

type Incident struct {
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

func (u *Incident) String() string {
	return fmt.Sprintf("Incident(%d, %s)", u.ID, u.Message)
}

func (u *Incident) Validate(v *revel.Validation) {
	v.Required(u.Message)

	v.MacAddr(u.Mac)
}
