package models

import (
	"time"

	gorp "gopkg.in/gorp.v2"
)

type OptOut struct {
	ID            int
	CreatedString string
	Mac           string
	Pesel         string
	Username      string
	Action        string
	Comment       string

	// transient
	Created time.Time
}

func (o *OptOut) PreInsert(_ gorp.SqlExecutor) error {
	o.CreatedString = o.Created.Format(time.RFC3339)

	return nil
}

func (o *OptOut) PostGet(_ gorp.SqlExecutor) error {
	t, err := time.Parse(time.RFC3339, o.CreatedString)
	if err != nil {
		return err
	}
	o.Created = t
	return nil
}
