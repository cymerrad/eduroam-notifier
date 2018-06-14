package models

type NotifierSettings struct {
	Template []byte
}

type NotifierRule struct {
	ID          int
	On, Do, Tag string
}
