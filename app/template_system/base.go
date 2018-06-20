package template_system

import (
	"bytes"
	"eduroam-notifier/app/models"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

type T struct {
	templates        map[TemplateID]*template.Template
	actions          map[Action]TemplateID
	replaceWithField map[TemplateTag]Field
	replaceWithConst map[TemplateTag]ConstValue
}

type TemplateID string
type Values map[string]string

var templateTags = regexp.MustCompile(`\{\{\s*(\w+)\s*\}\}`)

func New(settings models.NotifierSettings, rules []models.NotifierRule, templates []models.NotifierTemplate) (*T, error) {
	t := &T{}

	ts, err := ParseTemplates(templates)
	if err != nil {
		return nil, err
	}

	a, f, c, err := ParseRules(rules)
	if err != nil {
		return nil, err
	}

	t.templates, t.actions, t.replaceWithField, t.replaceWithConst = ts, a, f, c

	return t, nil
}

func ParseTemplates(templates []models.NotifierTemplate) (out map[TemplateID]*template.Template, err error) {
	for _, tm := range templates {
		tmID := TemplateID(strconv.Itoa(tm.ID))

		// well crap, I totally forgot about how powerfull Golang's templating is
		tmBody := string(tm.Body)
		tmBodyDotted := templateTags.ReplaceAllString(tmBody, "{{.$1}}")

		tmpl, err := template.New(string(tmID)).Parse(tmBodyDotted)
		if err != nil {
			// we don't want non-parseable templates
			return out, err
		}
		out[tmID] = tmpl
	}
	return out, err
}

func ParseRules(rules []models.NotifierRule) (outA map[Action]TemplateID, outF map[TemplateTag]Field, outC map[TemplateTag]ConstValue, err error) {
	for _, rl := range rules {
		values := Values{}
		err := json.NewDecoder(strings.NewReader(rl.Value)).Decode(&values)
		// TODO: do something on all the errors below
		if err != nil {
			// error parsing
			continue
		}

		switch rl.On {
		case OnAction:
			switch rl.Do {
			case DoActionSendTemplate:
				action := values[OnAction]
				templateID := values[DoActionSendTemplate]
				outA[Action(action)] = TemplateID(templateID)

			default:
				// unrecognized
				continue
			}
			continue
		case OnTemplateTag:
			switch rl.Do {
			case DoInsertText:
				tag := TemplateTag(values[OnTemplateTag])
				constValue := ConstValue(values[DoInsertText])
				outC[tag] = constValue

			case DoSubstituteWithField:
				tag := TemplateTag(values[OnTemplateTag])
				field := Field(values[DoSubstituteWithField])
				outF[tag] = field

			default:
				// unrecognized
				continue
			}
		default:
			// unrecognized option
			continue
		}
	}

	return
}

func (t *T) Input(fieldsStruct models.EventMessageFields) (string, error) {
	var fieldsMap map[string]string
	btz, _ := json.Marshal(fieldsStruct)
	json.Unmarshal(btz, &fieldsMap)

	data := make(map[string]string)
	// gather data from fieldsStruct
	for key, value := range t.replaceWithField {
		data[string(key)] = string(fieldsMap[string(value)])
	}
	// throw in the rest
	for key, value := range t.replaceWithConst {
		data[string(key)] = string(value)
	}

	// this will be the output
	out := new(bytes.Buffer)

	// get the template we need
	tmplID := t.actions[Action(fieldsStruct.Action)]
	tmpl := t.templates[tmplID]

	// execute template
	tmpl.Execute(out, data)

	return out.String(), nil
}
