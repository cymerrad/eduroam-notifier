package models

type NotifierSettings struct {
	Something string
	Template  []byte
}

type NotifierRule struct {
	ID          int
	On, Do, Tag string
}
