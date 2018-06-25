package models

import (
	"encoding/json"
	"time"
)

type NotifierSettings struct {
	ID      int
	JSON    []byte
	Created time.Time
}

func (ns NotifierSettings) Unmarshall() (NotifierSettingsParsed, error) {
	settingsParsed := NotifierSettingsParsed{}
	err := json.Unmarshal(ns.JSON, &settingsParsed)
	return settingsParsed, err
}

type NotifierSettingsParsed struct {
	Cooldown int64 `json:"cooldown"`
}

type NotifierTemplate struct {
	ID      int
	Body    []byte
	Created time.Time
}

type NotifierRule struct {
	ID            int
	On, Do, Value string
	Created       time.Time
}
