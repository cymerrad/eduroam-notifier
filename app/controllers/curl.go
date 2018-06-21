package controllers

import (
	"eduroam-notifier/app/models"
	"eduroam-notifier/app/template_system"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/revel/revel"
	sq "gopkg.in/Masterminds/squirrel.v1"
)

type Curl struct {
	App
}

type CurlData struct {
	Input, Output string
}

type SettingsData struct {
	Templates    []BodyParsed
	TemplatesRaw []models.NotifierTemplate
	Rules        []models.NotifierRule
	OtherParsed  models.NotifierSettingsParsed
	Other        string

	Schema string
}

type BodyParsed struct {
	ID   int
	Body string
}

func (c Curl) Index() revel.Result {
	if c.Validation.HasErrors() {
		return c.Render()
	}
	settings, err := c.retrieveSettingsFromDB()
	if err != nil {
		c.Validation.Error("Error occurred: %s", err.Error())
	}
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
	}

	c.ViewArgs["settings"] = settings

	return c.Render()
}

func (c Curl) HasErrorsRedirect(val interface{}) (res revel.Result, ok bool) {
	if c.Validation.HasErrors() {
		// Store the validation errors in the flash context and redirect.
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Curl.Index), true
	}
	return revel.ErrorResult{}, false
}

func (c Curl) Notify() revel.Result {
	rawJSON := c.Params.Get("json")
	input := make(map[string]interface{})
	err := json.NewDecoder(strings.NewReader(rawJSON)).Decode(&input)
	if err != nil {
		if strings.HasPrefix(err.Error(), `invalid character '\''`) {
			c.Validation.Error("Use double quotes instead of single quotes.")
		} else {
			c.Validation.Error("Parsing returned this: %s", err.Error())
		}
	}
	if res, ok := c.HasErrorsRedirect(Curl.Index); ok {
		return res
	}

	event := models.EventParsed{}
	_ = json.NewDecoder(strings.NewReader(rawJSON)).Decode(&event)

	c.Log.Debugf("Form: %#v", c.Params.Form)

	prettiedUpInput, _ := json.MarshalIndent(input, "", "  ")

	// creating temporary settings for testing purposes
	settings, err := c.retrieveSettingsFromSession()
	if err != nil {
		c.Validation.Error("Parsing returned this: %s", err.Error())
	}
	if res, ok := c.HasErrorsRedirect(Curl.Index); ok {
		return res
	}

	templates, err := template_system.New(settings.OtherParsed, settings.Rules, settings.TemplatesRaw)
	if err != nil {
		c.Validation.Error("Error occurred: %s", err.Error())
	}

	c.ViewArgs["curl"] = CurlData{
		Input:  string(prettiedUpInput),
		Output: c.dryRun(event, templates),
	}
	if res, ok := c.HasErrorsRedirect(Curl.Index); ok {
		return res
	}

	c.ViewArgs["settings"] = settings

	// TODO frontend bugs
	c.Log.Debugf("Settings: %#v", settings)

	return c.RenderTemplate("Curl/Index.html")
}

func (c App) retrieveSettingsFromDB() (s SettingsData, err error) {
	txn := c.Txn
	var templatesRaw []models.NotifierTemplate
	str, _, _ := sq.StatementBuilder.Select("*").From("NotifierTemplate").ToSql()
	_, _ = txn.Select(&templatesRaw, str)

	var rules []models.NotifierRule
	str2, _, _ := sq.StatementBuilder.Select("*").From("NotifierRule").ToSql()
	_, _ = txn.Select(&rules, str2)

	settings := models.NotifierSettings{}
	str3 := "SELECT * FROM NotifierSettings WHERE ID = ( SELECT MAX(ID) FROM NotifierSettings ) LIMIT 1"
	err = txn.SelectOne(&settings, str3)
	if err != nil {
		return s, errors.New("no settings")
	}

	templatesParsed := make([]BodyParsed, len(templatesRaw))
	for ind, raw := range templatesRaw {
		templatesParsed[ind] = BodyParsed{raw.ID, string(raw.Body)}
	}

	schemaParsed, _ := json.Marshal(template_system.Schema)
	settingsParsed, _ := settings.Unmarshall()

	return SettingsData{
		Templates:    templatesParsed,
		Rules:        rules,
		OtherParsed:  settingsParsed,
		Schema:       string(schemaParsed),
		Other:        string(settings.JSON),
		TemplatesRaw: templatesRaw,
	}, nil
}

func (c App) retrieveSettingsFromSession() (s SettingsData, err error) {
	otherRaw := c.Params.Get("other")
	other := models.NotifierSettingsParsed{}
	err = json.NewDecoder(strings.NewReader(otherRaw)).Decode(&other)
	if err != nil {
		return s, err
	}

	cases := c.Params.Values["settings-cases"]
	rules, err := template_system.ParseRulesFromValues(cases)
	if err != nil {
		return s, err
	}
	templatesRaw := []models.NotifierTemplate{}
	key := templateKey{1}
	for {
		if val := c.Params.Get(key.Get()); val != "" {
			templatesRaw = append(templatesRaw, models.NotifierTemplate{
				Body: []byte(val),
				ID:   key.ID(),
			})
			key.Next()
		} else {
			break
		}
	}
	templatesPrettied := make([]BodyParsed, len(templatesRaw))
	for ind, tmpl := range templatesRaw {
		templatesPrettied[ind] = BodyParsed{
			ID:   tmpl.ID,
			Body: string(tmpl.Body),
		}
	}

	schemaParsed, _ := json.Marshal(template_system.Schema)

	settings := SettingsData{
		OtherParsed:  other,
		Other:        otherRaw,
		Rules:        rules,
		TemplatesRaw: templatesRaw,
		Templates:    templatesPrettied,
		Schema:       string(schemaParsed),
	}

	return settings, err
}

func (c Curl) dryRun(event models.EventParsed, template *template_system.T) string {
	out := strings.Builder{}

	for _, match := range event.CheckResult.MatchingMessages {
		result := interpretMessage(match.Fields, template)
		btz, _ := json.MarshalIndent(result, "", "  ")
		out.WriteString(string(btz) + "\n")
	}

	out.WriteString(witnessMeBloodBag + "\n")

	return out.String()
}

const witnessMeBloodBag = `__        ___ _                       _ 
\ \      / (_) |_ _ __   ___  ___ ___| |
 \ \ /\ / /| | __| '_ \ / _ \/ __/ __| |
  \ V  V / | | |_| | | |  __/\__ \__ \_|
   \_/\_/  |_|\__|_| |_|\___||___/___(_)
                                        `

type templateKey struct {
	id int
}

func (tk *templateKey) Next() {
	tk.id++
}
func (tk *templateKey) Get() string {
	return fmt.Sprintf("template%d", tk.id)
}
func (tk *templateKey) ID() int {
	return tk.id
}
