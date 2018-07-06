package controllers

import (
	"crypto/sha256"
	"eduroam-notifier/app/mailer"
	"eduroam-notifier/app/models"
	ts "eduroam-notifier/app/ts"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/revel/revel"
	gorp "gopkg.in/gorp.v2"
)

type Notifier struct {
	App
	mailer mailer.M
}

var USOSdbm *gorp.DbMap
var Mailer *mailer.M

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

	msgs, err := c.interpretEvent(parsedEvent, event.ID, globalTemplate)
	if err != nil {
		c.Log.Errorf("Interpreting: %s", err.Error())
	} else {
		c.Log.Debugf("Result %v", msgs)
	}

	return c.RenderJSON(msgs)
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

	schemaParsed, _ := json.Marshal(ts.Schema)
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
	rules, err := ts.ParseRulesFromValues(cases)
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

	schemaParsed, _ := json.Marshal(ts.Schema)

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

func (c Notifier) interpretEvent(event models.EventParsed, eventID int, templateSystem *ts.T) ([]models.MailMessage, error) {
	out := make([]models.MailMessage, 0)

	for _, match := range event.CheckResult.MatchingIncidents {
		extras := make(map[string]string)
		incid := match.ToIncident(0)

		// price of using goto
		var body string
		var subject string
		var err error
		var countMsgs int64
		var emailAddr string
		var ignoreFirst int
		var previous int64
		var otherData OtherUserData
		// var otherDataMap map[string]string

		if err = c.Txn.Insert(&incid); err != nil {
			c.Log.Errorf("Saving the EVENT Incident: %s", err.Error())
		}

		mailMsg := models.MailMessage{}
		stampTheMessage(&mailMsg, &match.Fields, eventID)

		optOuts := make([]models.OptOut, 0)

		// PRE-RENDER CHECKS
		_, err = c.Txn.Select(&optOuts, models.GetOptOutsOfUser(incid))
		if err != nil {
			c.Log.Errorf("Opt-out finding failed. Refusing to take action. %s", err.Error())
			return nil, err
		}
		if len(optOuts) > 0 {
			c.Log.Debugf("Opt-out: %#v", optOuts)
			mailMsg.Error = "User opted-out from notifications."

			goto SKIPPING_STUFF
		}

		ignoreFirst, err = templateSystem.Preflight(match.Fields)
		if err != nil {
			c.Log.Errorf("POSSIBLY UNRECOGNIZED ACTION: %#v", match)
			mailMsg.Error = fmt.Sprintf("Preflight error on action '%s'.", match.Fields.Action)

			goto SKIPPING_STUFF
		}

		previous, err = c.Txn.SelectInt(models.GetCountIncidentsLikeByMac(incid))
		if err != nil {
			c.Log.Errorf("Counting previous failed.")
			return nil, err
		}
		if previous < int64(ignoreFirst) {
			mailMsg.Error = fmt.Sprintf("TOO EARLY FOR SPAMMING: prev %d < ignore %d", previous, ignoreFirst)

			goto SKIPPING_STUFF
		}

		// CONSTANTS FOR TEMPLATING
		emailAddr, err = getUserEmailAddress(&match.Fields)
		if err != nil {
			c.Log.Errorf("getUserEmailAddress: %s", err.Error())
			extras[ts.CANCEL_LINK] = "(could not be generated, sorry)"
		} else {
			hash := ConvertEmailAddressToHash(emailAddr)
			urlPath, err := revel.ReverseURL("Notifier.Cancel", hash)
			clickyLink := fmt.Sprintf("http://%s%s", c.Request.Host, urlPath)
			// clickyLink := fmt.Sprintf("<a href=\"%s\">Click me</a>", urlFull)

			if err != nil {
				c.Log.Errorf("Generating link: %s", err.Error())
				extras[ts.CANCEL_LINK] = "(could not be generated, sorry)"
			} else {
				extras[ts.CANCEL_LINK] = clickyLink
			}
		}
		otherData, err = getOtherUserData(&match.Fields)
		if err != nil {
			c.Log.Errorf("getOtherUserData: %s", err.Error())
		} else {
			otherDataMap := otherData.ToMap()
			for k, v := range otherDataMap {
				extras[k] = v
			}
		}

		countMsgs, err = c.Txn.SelectInt(models.GetCountIncidentsLikeByMac(incid))
		if err != nil {
			c.Log.Errorf("Executing counting query: %s", err.Error())
		} else {
			extras[ts.COUNT_MAC] = strconv.FormatInt(countMsgs, 10)
		}

		countMsgs, err = c.Txn.SelectInt(models.GetCountIncidentsLikeByPesel(incid))
		if err != nil {
			c.Log.Errorf("Executing counting query: %s", err.Error())
		} else {
			extras[ts.COUNT_PESEL] = strconv.FormatInt(countMsgs, 10)
		}

		countMsgs, err = c.Txn.SelectInt(models.GetCountIncidentsLikeByUsername(incid))
		if err != nil {
			c.Log.Errorf("Executing counting query: %s", err.Error())
		} else {
			extras[ts.COUNT_USERNAME] = strconv.FormatInt(countMsgs, 10)
		}

		// FINALLY GET THE CONTENTS OF THE Incident
		body, subject, err = interpretIncident(match.Fields, extras, templateSystem)
		if err != nil {
			c.Log.Errorf("Generating body: %s", err.Error())
			mailMsg.Error = err.Error()
		}

		mailMsg.BodyString = body
		mailMsg.Subject = subject

		// IF ANYTHING BAD HAPPEND DURING THIS PROCEDURE, WE WILL ARRIVE HERE - SKIPPING GENERATING BODY ET AL.
	SKIPPING_STUFF:

		out = append(out, mailMsg)

		if err := c.Txn.Insert(&mailMsg); err != nil {
			c.Log.Errorf("Saving the MAIL Incident: %s", err.Error())
		}
	}

	return out, nil
}

func interpretIncident(fields models.EventIncidentFields, extras map[string]string, a *ts.T) (string, string, error) {
	var fieldsMap map[string]string
	btz, _ := json.Marshal(fields)
	err := json.Unmarshal(btz, &fieldsMap)
	if err != nil {
		revel.AppLog.Errorf("fieldsStruct -> fieldsMap error: %s", err.Error())
		return "", "", err
	}

	action := fields.Action

	// add some missing fields
	fieldsMap["gl2_remote_port"] = strconv.Itoa(fields.Gl2RemotePort)
	fieldsMap["level"] = strconv.Itoa(fields.Level)

	return a.Input(action, fieldsMap, extras)
}

// TODO
// this will seriously change (e.g. make calls to some other service)
func getUserEmailAddress(fields *models.EventIncidentFields) (recipient string, err error) {
	pesel := fields.Pesel
	emailAddr := ""
	err = USOSdbm.SelectOne(&emailAddr, "SELECT EMAIL FROM DZ_OSOBY WHERE PESEL=?", pesel)
	if err != nil {
		recipient = "(cannot be found)"
	} else {
		recipient, err = emailAddr, nil
	}
	return
}

var otherUserDataColumns = "imie, imie2, NAZWISKO, PLEC, NAR_KOD"

type OtherUserData struct {
	Imie     string `json:"FIRST_NAME"`
	Imie2    string `json:"SECOND_NAME"`
	Nazwisko string `json:"SURNAME"`
	Plec     string `json:"SEX"`
	Nar_Kod  string `json:"NATIONALITY"`
}

func (o OtherUserData) ToMap() map[string]string {
	there, _ := json.Marshal(o)
	andBackAgain := make(map[string]string)
	_ = json.Unmarshal(there, &andBackAgain)
	return andBackAgain
}

func getOtherUserData(fields *models.EventIncidentFields) (OtherUserData, error) {
	var data OtherUserData
	pesel := fields.Pesel

	err := USOSdbm.SelectOne(&data, "SELECT "+otherUserDataColumns+" FROM DZ_OSOBY WHERE PESEL=?", pesel)
	return data, err
}

func stampTheMessage(incid *models.MailMessage, fields *models.EventIncidentFields, eventID int) {
	recipient, err := getUserEmailAddress(fields)
	if err != nil {
		incid.Error = err.Error()
	}

	incid.Recipient = recipient
	incid.Created = time.Now()
	incid.EventID = eventID
	incid.Hash = ConvertEmailAddressToHash(recipient)
	incid.Pesel = fields.Pesel
}

func (c Notifier) Cancel(id string) revel.Result {
	// TODO
	// render pages for passing in the comments
	c.Log.Debugf("Requested cancellation of %s", id)

	var lastMsg models.Incident
	err := c.Txn.SelectOne(&lastMsg, models.GetLastIncidentByHash(id))
	if err != nil {
		c.Log.Errorf("Cancelling with hash: %s", err.Error())
		return c.RenderText("fial")
	}

	optOut, _ := CreateOptOutEntry(lastMsg, "This service sucks, that's why.")
	err = c.Txn.Insert(&optOut)
	if err != nil {
		c.Log.Errorf("Cancelling with hash, db query: %s", err.Error())
		return c.RenderText("fial")
	}

	return c.RenderText("k")
}

func ConvertEmailAddressToHash(email string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(email)))
}

func CreateOptOutEntry(filter models.Incident, comment string) (models.OptOut, error) {
	optOut := models.OptOut{
		Action:   filter.Action,
		Comment:  comment,
		Created:  time.Now(),
		Mac:      filter.Mac,
		Pesel:    filter.Pesel,
		Username: filter.Username,
	}

	return optOut, nil
}
