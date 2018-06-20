package template_system

type Field string
type TemplateTag string

const (
	OnTemplateTag = "template_tag"
)

const (
	DoSubstituteWithField = "substitute_with_field"
	DoInsertText          = "insert_text"
)
