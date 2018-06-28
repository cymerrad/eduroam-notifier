package ts

import (
	"eduroam-notifier/app/models"
	"time"
)

var TimeZero = time.Unix(0, 0)
var GoodStartingSettings = []models.NotifierRule{
	{
		On:      OnTemplateTag,
		Do:      DoInsertText,
		Value:   GenerateJSON(OnTemplateTag, "signature", DoInsertText, "DSK UW"),
		Created: TimeZero,
	},
	{
		On:      OnTemplateTag,
		Do:      DoSubstituteWithField,
		Value:   GenerateJSON(OnTemplateTag, "mac", DoSubstituteWithField, "source-mac"),
		Created: TimeZero,
	},
	{
		On:      OnTemplateTag,
		Do:      DoSubstituteWithField,
		Value:   GenerateJSON(OnTemplateTag, "pesel", DoSubstituteWithField, "Pesel"),
		Created: TimeZero,
	},
	{
		On:      OnAction,
		Do:      DoActionPickTemplate,
		Value:   GenerateJSON(OnAction, "Login incorrect (mschap: MS-CHAP2-Response is incorrect)", DoActionPickTemplate, "wrong_password"),
		Created: TimeZero,
	},
	{
		On:      OnAction,
		Do:      DoIgnoreFirstN,
		Value:   GenerateJSON(OnAction, "Login incorrect (mschap: MS-CHAP2-Response is incorrect)", DoIgnoreFirstN, "5"),
		Created: TimeZero,
	},
}
