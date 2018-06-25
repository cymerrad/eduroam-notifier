package models

import (
	"encoding/json"
	"time"

	gorp "gopkg.in/gorp.v2"
)

type NotifierSettings struct {
	ID            int
	JSON          []byte
	CreatedString string

	// transient
	Created time.Time
}

func (o *NotifierSettings) PreInsert(_ gorp.SqlExecutor) error {
	o.CreatedString = o.Created.Format(time.RFC3339)

	return nil
}

func (o *NotifierSettings) PostGet(_ gorp.SqlExecutor) error {
	t, err := time.Parse(time.RFC3339, o.CreatedString)
	if err != nil {
		return err
	}
	o.Created = t

	return nil
}

func (ns NotifierSettings) Unmarshall() (NotifierSettingsParsed, error) {
	settingsParsed := NotifierSettingsParsed{}
	err := json.Unmarshal(ns.JSON, &settingsParsed)
	return settingsParsed, err
}

type NotifierSettingsParsed struct {
	Cooldown int64 `json:"cooldown"`
}

func (ns NotifierSettingsParsed) Marshall() ([]byte, error) {
	return json.Marshal(ns)
}

type NotifierTemplate struct {
	ID            int
	Body          []byte
	CreatedString string

	// transient
	Created time.Time
}

func (o *NotifierTemplate) PreInsert(_ gorp.SqlExecutor) error {
	o.CreatedString = o.Created.Format(time.RFC3339)

	return nil
}

func (o *NotifierTemplate) PostGet(_ gorp.SqlExecutor) error {
	t, err := time.Parse(time.RFC3339, o.CreatedString)
	if err != nil {
		return err
	}
	o.Created = t
	return nil
}

type NotifierRule struct {
	ID            int
	On, Do, Value string
	CreatedString string

	// transient
	Created time.Time
}

func (o *NotifierRule) PreInsert(_ gorp.SqlExecutor) error {
	o.CreatedString = o.Created.Format(time.RFC3339)

	return nil
}

func (o *NotifierRule) PostGet(_ gorp.SqlExecutor) error {
	t, err := time.Parse(time.RFC3339, o.CreatedString)
	if err != nil {
		return err
	}
	o.Created = t
	return nil
}
