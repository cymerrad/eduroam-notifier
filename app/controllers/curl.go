package controllers

import (
	"database/sql"
	"eduroam-notifier/app/models"
	"encoding/json"
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
	Templates []models.NotifierTemplate
	Rules     []models.NotifierRule
	Settings  string
}

func (c Curl) Index() revel.Result {
	if c.Validation.HasErrors() {
		return c.Render()
	}
	c.retrieveSettings()
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

	c.retrieveSettings()
	if c.Validation.HasErrors() {
		// Store the validation errors in the flash context and redirect.
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Curl.Index)
	}

	return c.RenderTemplate("Curl/Index.html")
}

func (c Curl) retrieveSettings() {
	var templates []models.NotifierTemplate
	str, _, err := sq.StatementBuilder.Select("*").From("NotifierTemplate").ToSql()
	if err != nil {
		c.Log.Errorf("Failed to build query")
		c.Validation.Error("Error retrieving templates.")
	} else {
		_, err = c.Txn.Select(&templates, str)
		if err != nil {
			if err != sql.ErrNoRows {
				c.Log.Errorf("Failed to retrieve templates: %s", err.Error())
			}
			c.Validation.Error("No templates.")
		}
	}

	var rules []models.NotifierRule
	str2, _, err := sq.StatementBuilder.Select("*").From("NotifierRule").ToSql()
	if err != nil {
		c.Log.Errorf("Failed to build query")
		c.Validation.Error("Error retrieving rules.")
	} else {
		_, err = c.Txn.Select(&rules, str2)
		if err != nil {
			if err != sql.ErrNoRows {
				c.Log.Errorf("Failed to retrieve rules %s", err.Error())
			}
			c.Validation.Error("No rules.")
		}
	}

	settings := models.NotifierSettings{}
	str3 := "SELECT * FROM NotifierSettings WHERE ID = ( SELECT MAX(ID) FROM NotifierSettings ) LIMIT 1"
	err = c.Txn.SelectOne(&settings, str3)
	if err != nil {
		if err != sql.ErrNoRows {
			c.Log.Errorf("Failed to retrieve settings: %s", err.Error())
		}
		c.Validation.Error("No settings.")
	}

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return
	}

	c.ViewArgs["templates"] = SettingsData{
		Templates: templates,
		Rules:     rules,
		Settings:  string(settings.JSON),
	}
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
