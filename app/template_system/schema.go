package ts

var Schema map[string][]string = map[string][]string{
	OnAction:      {DoActionPickTemplate},
	OnTemplateTag: {DoSubstituteWithField, DoInsertText},
}
