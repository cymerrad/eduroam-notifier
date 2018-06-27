package ts

var Schema map[string][]string = map[string][]string{
	OnAction:      {DoActionPickTemplate, DoIgnoreFirstN},
	OnTemplateTag: {DoSubstituteWithField, DoInsertText},
}
