package ts

type Action string

const (
	OnAction = "action"
)

const (
	DoActionPickTemplate = "pick_template"
	DoActionIgnoreFirstN = "ignore_first_n"
	DoActionEnterSubject = "enter_subject"
)

const (
	DefaultAction = Action("*")
)
