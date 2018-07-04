package ts

import (
	"bytes"
	"eduroam-notifier/app/models"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
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

var (
	ErrDeclaredValueMismatch = errors.New("declared/value mismatch")
	ErrUnrecognizedOption    = func(in string) error { return fmt.Errorf("unrecognized option: %s", in) }
)

func ParseRules(rules []models.NotifierRule) (outA map[Action]TemplateID, outF map[TemplateTag]Field, outC map[TemplateTag]ConstValue, outI map[Action]int, err error) {
	outA = make(map[Action]TemplateID)
	outF = make(map[TemplateTag]Field)
	outC = make(map[TemplateTag]ConstValue)
	outI = make(map[Action]int)

	var STOP = false

	var ifNotOkBail = func(isIt bool) {
		if !isIt {
			err = ErrDeclaredValueMismatch
			STOP = true
		}
	}

	var extract = func(rl models.NotifierRule) error {
		values := Values{}
		err := json.NewDecoder(strings.NewReader(rl.Value)).Decode(&values)
		if err != nil {
			return err
		}

		switch rl.On {
		case OnAction:
			action, ok := values[OnAction]
			ifNotOkBail(ok)

			switch rl.Do {
			case DoActionPickTemplate:
				templateID, ok := values[DoActionPickTemplate]
				ifNotOkBail(ok)
				outA[Action(action)] = TemplateID(templateID)

			case DoIgnoreFirstN:
				ignoreValue, ok := values[DoIgnoreFirstN]
				ifNotOkBail(ok)
				parsed, err := strconv.Atoi(ignoreValue)
				if err != nil {
					return err
				}
				outI[Action(action)] = parsed

			default:
				// unrecognized
				return ErrUnrecognizedOption(rl.Do)
			}

		case OnTemplateTag:
			tag, ok := values[OnTemplateTag]
			ifNotOkBail(ok)

			switch rl.Do {
			case DoInsertText:
				constValue, ok := values[DoInsertText]
				ifNotOkBail(ok)
				outC[TemplateTag(tag)] = ConstValue(constValue)

			case DoSubstituteWithField:
				field, ok := values[DoSubstituteWithField]
				ifNotOkBail(ok)
				outF[TemplateTag(tag)] = Field(field)

			default:
				// unrecognized
				return ErrUnrecognizedOption(rl.Do)
			}
		default:
			// unrecognized option
			return ErrUnrecognizedOption(rl.On)
		}

		return nil
	}

	for _, rl := range rules {
		err2 := extract(rl)
		if STOP {
			return nil, nil, nil, nil, err
		}
		if err2 != nil {
			return nil, nil, nil, nil, err2
		}
	}

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

func (t *T) getTemplateIDOrDefault(in string) (TemplateID, error) {
	first, ok := t.Actions[Action(in)]
	if ok {
		return first, nil
	}
	second, ok := t.Actions[DefaultAction]
	if ok {
		return second, nil
	}
	return "", errors.New("no such action " + in + " and no default set")
}

//Preflight says how many first occurences to ignore and if there are any critical errors with the event.
func (t *T) Preflight(fieldsStruct models.EventIncidentFields) (int, error) {
	_, err := t.getTemplateIDOrDefault(fieldsStruct.Action)
	if err != nil {
		return 0, err
	}

	action := Action(fieldsStruct.Action)
	ignoreFirst, ok := t.IgnoreFirst[DefaultAction]
	if !ok {
		ignoreFirst = t.IgnoreFirst[action]
	}
	return ignoreFirst, nil
}

func (t *T) Input(fieldsStruct models.EventIncidentFields, extras map[string]string) (string, error) {
	// get the template we need
	tmplID, err := t.getTemplateIDOrDefault(fieldsStruct.Action)
	if err != nil {
		return "", err
	}

	tmpl, ok := t.Templates[tmplID]
	if !ok {
		return "", errors.New("no such template " + string(tmplID))
	}

	var fieldsMap map[string]string
	btz, _ := json.Marshal(fieldsStruct)
	err = json.Unmarshal(btz, &fieldsMap)
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
