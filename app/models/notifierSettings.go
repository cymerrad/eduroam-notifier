package models

type NotifierSettings struct {
	Something string
}

type NotifierRule struct {
	ID       int
	On, Do   string
	Template []byte
}
