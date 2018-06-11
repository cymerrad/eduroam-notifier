package controllers

import (
	"encoding/json"
	"strings"

	"github.com/revel/revel"
)

type Curl struct {
	App
}

type curlData struct {
	Input, Output string
}

func (c Curl) Index() revel.Result {
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
			c.Validation.Error(err.Error())
		}
	}

	if c.Validation.HasErrors() {
		// Store the validation errors in the flash context and redirect.
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Curl.Index)
	}

	prettiedUp, _ := json.MarshalIndent(input, "", "  ")

	c.ViewArgs["curl"] = curlData{
		Input:  string(prettiedUp),
		Output: c.dryRun(rawJSON),
	}

	return c.RenderTemplate("Curl/Index.html")
}

func (c Curl) dryRun(rawJSON string) string {

	return `__        ___ _                       _ 
\ \      / (_) |_ _ __   ___  ___ ___| |
 \ \ /\ / /| | __| '_ \ / _ \/ __/ __| |
  \ V  V / | | |_| | | |  __/\__ \__ \_|
   \_/\_/  |_|\__|_| |_|\___||___/___(_)
                                        `
}
