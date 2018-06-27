package controllers

import (
	"eduroam-notifier/app/models"
	"eduroam-notifier/app/template_system"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/revel/revel"
)

type Notifier struct {
	App
}

func (c Notifier) Notify() revel.Result {
	now := time.Now()

	parsedEvent, err := c.parseEvent()
	if err != nil {
		c.Log.Error("Error parsing event")
		c.Response.Status = http.StatusNotAcceptable
		return c.RenderText(err.Error())
	}

	event := models.Event{
		Body:    c.Params.JSON,
		Created: now,
	}
	if err := c.Txn.Insert(&event); err != nil {
		c.Log.Errorf("Error inserting event into DB: %s", err.Error())
		c.Response.Status = http.StatusInternalServerError
		return c.RenderText(":c")
	}

	for _, match := range parsedEvent.CheckResult.MatchingMessages {
		msg := match.ToMessage(event.ID)

		// store the match in database for further decisions
		if err := c.Txn.Insert(&msg); err != nil {
			c.Log.Errorf("Error inserting event into DB: %s", err.Error())
			continue
		}

		// TODO
		// at this point perform check if spamming the user is necessary
		// because we still have access to database at this point

		c.Log.Debugf("Success inserting message %v", msg)

		interpretMessage(match.Fields, nil, event.ID, globalTemplate)

	}

	return c.RenderText("k")
}

func (c Notifier) parseEvent() (models.EventParsed, error) {
	eventP := models.EventParsed{}
	err := c.Params.BindJSON(&eventP)
	return eventP, err
}

func (c Notifier) Settings() revel.Result {
	s, err := c.retrieveSettingsFromSession()
	if err != nil {
		c.Validation.Error("Corrupted form %s", err.Error())
	} else {
		err2 := c.saveSettings(s)
		if err != nil {
			c.Validation.Error("Saving settings failed %s", err2.Error())
		}
	}
	if res, ok := c.HasErrorsRedirect(Curl.Index); ok {
		return res
	}

	c.Log.Debugf("Form: %v", c.Params.Values)

	redirectTo := c.Params.Get("redirect")
	if redirectTo == "curl" {
		return c.Redirect(Curl.Index)
	}

	return c.Redirect(App.Index)
}

func (c Notifier) HasErrorsRedirect(val interface{}) (res revel.Result, ok bool) {
	if c.Validation.HasErrors() {
		// Store the validation errors in the flash context and redirect.
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Curl.Index), true
	}
	return revel.ErrorResult{}, false
}

func (c Notifier) saveSettings(s SettingsData) error {
	now := time.Now()

	btz, _ := s.OtherParsed.Marshall()
	settings := &models.NotifierSettings{
		Created: now,
		JSON:    btz,
	}

	errors := make([]error, 0)

	// copy, update created time and insert
	for _, el := range s.Rules {
		temp := models.NotifierRule(el)
		temp.Created = now
		errors = append(errors, c.Txn.Insert(&temp))
	}
	for _, el := range s.TemplatesRaw {
		temp := models.NotifierTemplate(el)
		temp.Created = now
		errors = append(errors, c.Txn.Insert(&temp))
	}
	errors = append(errors, c.Txn.Insert(settings))

	// TODO one error to rule them all?
	func(errs ...error) {
		for _, e := range errs {
			if e != nil {
				c.Log.Errorf("Error adding new settings: %s", e.Error())
			}
		}
	}(errors...)

	return nil
}

func (c Notifier) retrieveSettingsFromDB() (s SettingsData, err error) {
	txn := c.Txn
	var templatesRaw []models.NotifierTemplate
	_, _ = txn.Select(&templatesRaw, models.GetAllNotifierTemplates)

	var rules []models.NotifierRule
	_, _ = txn.Select(&rules, models.GetAllNotifierRules)

	settings := models.NotifierSettings{}
	err = txn.SelectOne(&settings, models.GetNotifierSettings)
	if err != nil {
		return s, errors.New("no settings")
	}

	templatesParsed := make([]BodyParsed, len(templatesRaw))
	for ind, raw := range templatesRaw {
		templatesParsed[ind] = BodyParsed{raw.ID, raw.Name, string(raw.Body)}
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

func (c Notifier) retrieveSettingsFromSession() (s SettingsData, err error) {
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
	keys := getAllTemplateKeys(c.Params.Values)
	c.Log.Debugf("KEYS: %v", keys)
	for k, v := range keys {
		if val := c.Params.Get(k); val != "" {
			templatesRaw = append(templatesRaw, models.NotifierTemplate{
				Body: []byte(val),
				Name: v,
			})
		}
	}
	templatesPrettied := make([]BodyParsed, len(templatesRaw))
	for ind, tmpl := range templatesRaw {
		templatesPrettied[ind] = BodyParsed{
			ID:   tmpl.ID,
			Name: tmpl.Name,
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

func (c App) interpretEvent(event models.EventParsed, eventID int, templateSystem *template_system.T) ([]models.MailMessage, error) {
	out := make([]models.MailMessage, 0)

	for _, match := range event.CheckResult.MatchingMessages {
		extras := make(map[string]string)
		msg := match.ToMessage(0)

		// PRE-RENDER CHECKS
		optOuts, err := c.Txn.Select(models.OptOut{}, models.GetOptOutsOfUser(msg))
		if err != nil {
			c.Log.Errorf("Opt-out finding failed. Refusing to take action.")
			return nil, err
		}
		if len(optOuts) > 0 {
			c.Log.Debugf("Opt-out: %#v", optOuts)
			mailMsg := models.MailMessage{}
			stampTheMessage(&mailMsg, &match.Fields, eventID)
			mailMsg.Error = "User opted-out from notifications."
			out = append(out, mailMsg)

			if err := c.Txn.Insert(&msg); err != nil {
				c.Log.Errorf("Saving the message: %s", err.Error())
			}

			return out, nil
		}

		// CONSTANTS FOR TEMPLATING
		countMsgs, err := c.Txn.SelectInt(models.GetCountMessagesLikeByMac(msg))
		if err != nil {
			c.Log.Errorf("Executing counting query: %s", err.Error())
		} else {
			extras["COUNT_MAC"] = strconv.FormatInt(countMsgs, 10)
		}

		countMsgs, err = c.Txn.SelectInt(models.GetCountMessagesLikeByPesel(msg))
		if err != nil {
			c.Log.Errorf("Executing counting query: %s", err.Error())
		} else {
			extras["COUNT_PESEL"] = strconv.FormatInt(countMsgs, 10)
		}

		countMsgs, err = c.Txn.SelectInt(models.GetCountMessagesLikeByUsername(msg))
		if err != nil {
			c.Log.Errorf("Executing counting query: %s", err.Error())
		} else {
			extras["COUNT_USERNAME"] = strconv.FormatInt(countMsgs, 10)
		}

		result := interpretMessage(match.Fields, extras, eventID, templateSystem)
		out = append(out, result)

		if err := c.Txn.Insert(&result); err != nil {
			c.Log.Errorf("Saving the message: %s", err.Error())
		}
	}

	return out, nil
}

func interpretMessage(fields models.EventMessageFields, extras map[string]string, eventID int, a *template_system.T) (resp models.MailMessage) {
	revel.AppLog.Debugf("Doing something magical with %#v", fields)

	stampTheMessage(&resp, &fields, eventID)

	output, err := a.Input(fields, extras)
	if err != nil {
		resp.Error += err.Error()
	} else {
		resp.Body = output
	}

	return
}

// TODO
// this might seriously change (e.g. make calls to some other service)
func stampTheMessage(msg *models.MailMessage, fields *models.EventMessageFields, eventID int) {
	var recipient string
	var err string

	if fields.SourceUser == "" {
		recipient, err = "(cannot be found)", "empty SourceUser field"
	} else {
		recipient, err = fields.SourceUser, ""
	}

	msg.Recipient = recipient
	msg.Error = err
	msg.Created = time.Now()
	msg.EventID = eventID
}
