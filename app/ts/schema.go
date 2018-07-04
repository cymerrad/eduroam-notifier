package ts

import "fmt"

var Schema map[string][]string = map[string][]string{
	OnAction:      {DoActionPickTemplate, DoActionIgnoreFirstN, DoActionEnterSubject},
	OnTemplateTag: {DoSubstituteWithField, DoInsertText},
}

func GenerateJSON(key1 string, value1 string, key2 string, value2 string) string {
	return fmt.Sprintf("{\"%s\" : \"%s\", \"%s\" : \"%s\"}", key1, value1, key2, value2)
}
