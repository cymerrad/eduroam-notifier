package template_system

var Schema map[string][]string = map[string][]string{
	OnAction:      {DoActionSendTemplate},
	OnTemplateTag: {DoSubstituteWithField, DoInsertText},
}
