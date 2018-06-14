package controllers

import (
	"eduroam-notifier/app/models"
	"encoding/json"
	"errors"
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
	Templates []BodyParsed
	Rules     []models.NotifierRule
	Settings  string
}

type BodyParsed struct {
	ID   int
	Body string
}

func (c Curl) Index() revel.Result {
	if c.Validation.HasErrors() {
		return c.Render()
	}
	err := c.retrieveSettings()
	if err != nil {
		c.Validation.Error("Error occurred: %s", err.Error())
	}
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
	}

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

	prettiedUp, _ := json.MarshalIndent(input, "", "  ")

	c.ViewArgs["curl"] = CurlData{
		Input:  string(prettiedUp),
		Output: c.dryRun(rawJSON),
	}

	err = c.retrieveSettings()
	if err != nil {
		c.Validation.Error("Error occurred: %s", err.Error())
	}

	if c.Validation.HasErrors() {
		// Store the validation errors in the flash context and redirect.
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Curl.Index)
	}

	return c.RenderTemplate("Curl/Index.html")
}

func (c Curl) retrieveSettings() error {
	var templatesRaw []models.NotifierTemplate
	str, _, _ := sq.StatementBuilder.Select("*").From("NotifierTemplate").ToSql()
	_, _ = c.Txn.Select(&templatesRaw, str)

	var rules []models.NotifierRule
	str2, _, _ := sq.StatementBuilder.Select("*").From("NotifierRule").ToSql()
	_, _ = c.Txn.Select(&rules, str2)

	settings := models.NotifierSettings{}
	str3 := "SELECT * FROM NotifierSettings WHERE ID = ( SELECT MAX(ID) FROM NotifierSettings ) LIMIT 1"
	err := c.Txn.SelectOne(&settings, str3)
	if err != nil {
		return errors.New("no settings")
	}

	templatesParsed := make([]BodyParsed, len(templatesRaw))
	for ind, raw := range templatesRaw {
		templatesParsed[ind] = BodyParsed{raw.ID, string(raw.Body)}
	}

	c.ViewArgs["settings"] = SettingsData{
		Templates: templatesParsed,
		Rules:     rules,
		Settings:  string(settings.JSON),
	}

	return nil
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
