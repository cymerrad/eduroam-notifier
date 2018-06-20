package template_system

import (
	"eduroam-notifier/app/models"
	"encoding/json"
	"text/template"
)

type T struct {
	templates        map[TemplateID]template.Template
	actions          map[Action]TemplateID
	replaceWithField map[TemplateTag]Field
	replaceWithConst map[TemplateTag]string
}

type Value string

type TemplateID string

func New(settings models.NotifierSettings, rules []models.NotifierRule, templates []models.NotifierTemplate) (*T, error) {
	a := &T{}

	return a, nil
}

func (t *T) Input(fieldsStruct models.EventMessageFields) (string, error) {
	var fieldsMap map[string]string
	btz, _ := json.Marshal(fieldsStruct)
	json.Unmarshal(btz, &fieldsMap)

	return "", nil
}
