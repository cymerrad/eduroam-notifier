package controllers

import (
	"eduroam-notifier/app/models"
	"eduroam-notifier/app/template_system"
	"encoding/json"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/revel/revel"
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
	c.ViewArgs["curl"] = map[string]string{"hello": "there"}

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

func (c Curl) dryRun(event models.EventParsed, template *template_system.T) string {
	out := make([]ResponseAction, 0)

	for _, match := range event.CheckResult.MatchingMessages {
		result := interpretMessage(match.Fields, template)
		out = append(out, result)
	}

	btz, _ := json.MarshalIndent(out, "", "  ")

	return string(btz)
}

const witnessMeBloodBag = `__        ___ _                       _ 
\ \      / (_) |_ _ __   ___  ___ ___| |
 \ \ /\ / /| | __| '_ \ / _ \/ __/ __| |
  \ V  V / | | |_| | | |  __/\__ \__ \_|
   \_/\_/  |_|\__|_| |_|\___||___/___(_)
                                        `

var templateRe = regexp.MustCompile(`^template(\d+)$`)

func getAllTemplateKeys(form url.Values) map[string]int {
	keys := make(map[string]int)
	for k, _ := range form {
		if templateRe.MatchString(k) {
			res := templateRe.FindStringSubmatch(k)
			if res != nil && len(res) > 1 {
				keys[k], _ = strconv.Atoi(res[1])
			}
		}
	}
	return keys
}
