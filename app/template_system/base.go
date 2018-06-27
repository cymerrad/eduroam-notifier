package ts

import (
	"bytes"
	"eduroam-notifier/app/models"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"text/template"

	"github.com/revel/revel"
)

type T struct {
	Templates        map[TemplateID]*template.Template
	Actions          map[Action]TemplateID
	ReplaceWithField map[TemplateTag]Field
	ReplaceWithConst map[TemplateTag]ConstValue
	IgnoreFirst      map[Action]int
}

type TemplateID string
type Values map[string]string

var templateTags = regexp.MustCompile(`\{\{\s*(\w+)\s*\}\}`)

func New(other models.NotifierSettingsParsed, rules []models.NotifierRule, templates []models.NotifierTemplate) (*T, error) {
	ts, err := ParseTemplates(templates)
	if err != nil {
		return nil, err
	}

	a, f, c, i, err := ParseRules(rules)
	if err != nil {
		return nil, err
	}

	t := T{
		ts, a, f, c, i,
	}

	return &t, nil
}

func ParseTemplates(templates []models.NotifierTemplate) (out map[TemplateID]*template.Template, err error) {
	out = make(map[TemplateID]*template.Template)

	for _, tm := range templates {
		tmID := TemplateID(tm.Name)

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

var DeclaredValueMismatch = errors.New("declared/value mismatch")

func ParseRules(rules []models.NotifierRule) (outA map[Action]TemplateID, outF map[TemplateTag]Field, outC map[TemplateTag]ConstValue, outI map[Action]int, err error) {
	outA = make(map[Action]TemplateID)
	outF = make(map[TemplateTag]Field)
	outC = make(map[TemplateTag]ConstValue)
	outI = make(map[Action]int)

	var ifNotOkBail = func(isIt bool) {
		if !isIt {
			err = DeclaredValueMismatch
			goto dupa
		}
	}

	for _, rl := range rules {
		values := Values{}
		err = json.NewDecoder(strings.NewReader(rl.Value)).Decode(&values)
		if err != nil {
			return
		}

		switch rl.On {
		case OnAction:
			switch rl.Do {
			case DoActionPickTemplate:
				// TODO what if these are empty? Do some error handling finally
				action, ok := values[OnAction]
				templateID := values[DoActionPickTemplate]
				outA[Action(action)] = TemplateID(templateID)

			case DoIgnoreFirstN:
				action := values[OnAction]
				ignoreValue := values[DoIgnoreFirstN]
				outI[Action(action)] = 0

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

dupa:
	return
}

// Misleading af, lol
func ParseRulesFromValues(rules []string) ([]models.NotifierRule, error) {
	out := make([]models.NotifierRule, len(rules))
	for ind, rl := range rules {
		values := Values{}
		err := json.NewDecoder(strings.NewReader(rl)).Decode(&values)
		if err != nil {
			// error parsing
			continue
		}

		rule := &models.NotifierRule{}
		// first
		for key := range Schema {
			if _, ok := values[key]; ok {
				rule.On = key
				break
			}
		}
		if rule.On == "" {
			// error parsing
			continue
		}

		// second
		for _, key := range Schema[rule.On] {
			if _, ok := values[key]; ok {
				rule.Do = key
				break
			}
		}
		if rule.Do == "" {
			// error parsing
			continue
		}

		rule.Value = rl
		rule.ID = ind

		out[ind] = *rule
	}
	return out, nil
}

//Preflight says how many first occurences to ignore and if there are any critical errors with the event.
func (t *T) Preflight(fieldsStruct models.EventMessageFields) (int, error) {
	action := Action(fieldsStruct.Action)

	tmplID, ok := t.Actions[action]
	if !ok {
		return 0, errors.New("no such action " + fieldsStruct.Action)
	}

	ignoreFirst := t.IgnoreFirst[action]
	return ignoreFirst, nil
}

func (t *T) Input(fieldsStruct models.EventMessageFields, extras map[string]string) (string, error) {
	// get the template we need
	tmplID := t.Actions[Action(fieldsStruct.Action)]
	tmpl, ok := t.Templates[tmplID]
	if !ok {
		return "", errors.New("no such template " + string(tmplID))
	}

	var fieldsMap map[string]string
	btz, _ := json.Marshal(fieldsStruct)
	err := json.Unmarshal(btz, &fieldsMap)
	if err != nil {
		revel.AppLog.Errorf("fieldsStruct -> fieldsMap error: %s", err.Error())
	}

	data := make(map[string]string)
	// gather data from fieldsStruct
	for key, value := range t.ReplaceWithField {
		data[string(key)] = string(fieldsMap[string(value)])
	}
	// throw in the rest
	for key, value := range t.ReplaceWithConst {
		data[string(key)] = string(value)
	}
	for key, value := range extras {
		data[key] = value
	}

	revel.AppLog.Debugf("fieldsMap %#v", fieldsMap)
	revel.AppLog.Debugf("fieldsStruct %#v", fieldsStruct)
	revel.AppLog.Debugf("Data for template: %#v", data)

	// this will be the output
	out := new(bytes.Buffer)

	// execute template
	tmpl.Execute(out, data)

	return out.String(), nil
}

func (t *T) Show() string {
	lol, _ := json.Marshal(t)
	return string(lol)
}
