package models

type NotifierSettings struct {
	ID   int
	JSON []byte
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
