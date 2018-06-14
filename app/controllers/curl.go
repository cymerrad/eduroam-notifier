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
	Template string
	Rules    []models.NotifierRule
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
	settings := models.NotifierSettings{}
	str, _, err := sq.StatementBuilder.Select("*").From("NotifierSettings").Limit(1).ToSql()
	if err != nil {
		c.Log.Errorf("Failed to build query")
		c.Validation.Error("Error retrieving settings.")
	} else {
		err = c.Txn.SelectOne(&settings, str) // why do I have to pass the 'username' second time?
		if err != nil {
			if err != sql.ErrNoRows {
				c.Log.Errorf("Failed to retrieve settings: %s", err.Error())
			}
			c.Validation.Error("No settings.")
		}
	}

	var rules []models.NotifierRule
	str2, _, err := sq.StatementBuilder.Select("*").From("NotifierRule").ToSql()
	if err != nil {
		c.Log.Errorf("Failed to build query")
		c.Validation.Error("Error retrieving rules.")
	} else {
		_, err = c.Txn.Select(&rules, str2) // why do I have to pass the 'username' second time?
		if err != nil {
			if err != sql.ErrNoRows {
				c.Log.Errorf("Failed to retrieve rules %s", err.Error())
			}
			c.Validation.Error("No rules.")
		}
	}

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return
	}

	c.ViewArgs["settings"] = SettingsData{
		Template: string(settings.Template),
		Rules:    rules,
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
