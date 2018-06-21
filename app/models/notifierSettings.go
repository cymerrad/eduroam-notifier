package models

import (
	"encoding/json"
)

type NotifierSettings struct {
	ID   int
	JSON []byte
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
	ID   int
	Body []byte
}

type NotifierRule struct {
	ID            int
	On, Do, Value string
}
