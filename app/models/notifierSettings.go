package models

type NotifierSettings struct {
	ID   int
	JSON []byte
}

type NotifierSettingsParsed struct {
	Cooldown int64
}

type NotifierTemplate struct {
	ID   int
	Body []byte
}

type NotifierRule struct {
	ID                 int
	Tag, On, Do, Value string
}
