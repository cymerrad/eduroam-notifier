package controllers

import (
	"eduroam-notifier/app/models"
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-gorp/gorp"
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
	Templates []BodyParsed
	Rules     []models.NotifierRule
	Settings  string
	Schema    string
}

type BodyParsed struct {
	ID   int
	Body string
}

func (c Curl) Index() revel.Result {
	if c.Validation.HasErrors() {
		return c.Render()
	}
	settings, err := retrieveSettings(c.Txn)
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

func (c Curl) Notify() revel.Result {
	rawJSON := c.Params.Get("json")
	input := make(map[string]interface{})
	err := json.NewDecoder(strings.NewReader(rawJSON)).Decode(&input)
	if err != nil {
		if strings.HasPrefix(err.Error(), `invalid character '\''`) {
			c.Validation.Error("Use double quotes instead of single quotes.")
		} else {
			c.Validation.Error("Parsing returned this: %s", err.Error()) // FIXME REVEL HAS SOME PROBLEM WITH THIS LINE WITHOUT THE STRING
		}
	}

	if c.Validation.HasErrors() {
		// Store the validation errors in the flash context and redirect.
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Curl.Index)
	}

	c.Log.Infof("Form: %#v", c.Params.Form)

	prettiedUp, _ := json.MarshalIndent(input, "", "  ")

	c.ViewArgs["curl"] = CurlData{
		Input:  string(prettiedUp),
		Output: c.dryRun(rawJSON),
	}

	settings, err := retrieveSettings(c.Txn)
	if err != nil {
		c.Validation.Error("Error occurred: %s", err.Error())
	}

	if c.Validation.HasErrors() {
		// Store the validation errors in the flash context and redirect.
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Curl.Index)
	}

	c.ViewArgs["settings"] = settings

	return c.RenderTemplate("Curl/Index.html")
}

func retrieveSettings(txn *gorp.Transaction) (s SettingsData, err error) {
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

	schemaParsed, _ := json.Marshal(schema)

	return SettingsData{
		Templates: templatesParsed,
		Rules:     rules,
		Settings:  string(settings.JSON),
		Schema:    string(schemaParsed),
	}, nil
}

func (c Curl) dryRun(rawJSON string) string {
	return witness
}

const witness = `__        ___ _                       _ 
\ \      / (_) |_ _ __   ___  ___ ___| |
 \ \ /\ / /| | __| '_ \ / _ \/ __/ __| |
  \ V  V / | | |_| | | |  __/\__ \__ \_|
   \_/\_/  |_|\__|_| |_|\___||___/___(_)
                                        `

var schema map[string][]string = map[string][]string{
	"action":       {"send_template"},
	"template_tag": {"substitute_with_field", "insert_text"},
}
