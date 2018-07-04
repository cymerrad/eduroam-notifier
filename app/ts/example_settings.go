package ts

import (
	"eduroam-notifier/app/models"
	"time"
)

var TimeZero = time.Unix(0, 0)
var StartingRules = []models.NotifierRule{
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
		Do:      DoActionIgnoreFirstN,
		Value:   GenerateJSON(OnAction, "Login incorrect (mschap: MS-CHAP2-Response is incorrect)", DoActionIgnoreFirstN, "5"),
		Created: TimeZero,
	},
	{
		On:      OnAction,
		Do:      DoActionEnterSubject,
		Value:   GenerateJSON(OnAction, "Login incorrect (mschap: MS-CHAP2-Response is incorrect)", DoActionEnterSubject, "Ostrzeżenie Eduroam"),
		Created: TimeZero,
	},
}

const exTemp = `Witam.
Użytkowniku o numerze pesel {{pesel}} próbowałeś zalogować się z urządzenia {{mac}}, ale wprowadziłeś złe hasło po raz {{COUNT_MAC}}.
Jeżeli nie chcesz otrzymywać więcej takich maili, kliknij w {{CANCEL_LINK}}.

Z poważaniem,
{{signature}}`

var StartingTemplate = models.NotifierTemplate{
	Name:    "wrong_password",
	Body:    []byte(exTemp),
	Created: TimeZero,
}

var StartingSettings = models.NotifierSettingsParsed{
	Cooldown: int64(7 * 24 * time.Hour),
}
