package models

type NotifierSettings struct {
}

type NotifierRule struct {
	ID       int
	On, Do   string
	Template []byte
}
