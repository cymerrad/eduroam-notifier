package models

type NotifierSettings struct {
	Rules []*NotifierRule
}

type NotifierRule struct {
	On, Do   string
	Template []byte
}
