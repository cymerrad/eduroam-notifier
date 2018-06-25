package controllers

import (
	"database/sql"
	"eduroam-notifier/app/models"
	"eduroam-notifier/app/routes"
	"eduroam-notifier/app/template_system"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"

	sq "gopkg.in/Masterminds/squirrel.v1"
)

type App struct {
	GorpController
}

var globalTemplate *template_system.T

func (c App) Index() revel.Result {
	if c.connected() != nil {
		return c.Redirect(routes.App.Console())
	}
	return c.Render()
}

func (c App) Console() revel.Result {
	// c.ViewArgs["settings"] = settings

	return c.Render()
}

func (c App) AddUser() revel.Result {
	if user := c.connected(); user != nil {
		c.ViewArgs["user"] = user
	}
	return nil
}

func (c App) connected() *models.User {
	if c.ViewArgs["user"] != nil {
		return c.ViewArgs["user"].(*models.User)
	}
	if username, ok := c.Session["user"]; ok {
		return c.getUser(username)
	}
	return nil
}

func (c App) checkUser() revel.Result {
	if user := c.connected(); user == nil {
		c.Flash.Error("Please log in first")
		return c.Redirect(routes.App.Index())
	}
	return nil
}

func (c App) getUser(username string) (user *models.User) {
	user = &models.User{}

	str, _, err := sq.StatementBuilder.Select("*").From("User").Where(sq.Eq{"Username": username}).ToSql()
	if err != nil {
		c.Log.Errorf("Failed to build query")
		return nil
	}
	err = c.Txn.SelectOne(user, str, username) // why do I have to pass the 'username' second time?
	if err != nil {
		c.Log.Debugf("Failed query: %s; (%v)", str, user)
		if err != sql.ErrNoRows {
			c.Log.Error("Failed to find user")
		}
		return nil
	}
	return
}

func (c App) Login(username, password string, remember bool) revel.Result {
	user := c.getUser(username)
	if user != nil {
		err := bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password))
		if err == nil {
			c.Session["user"] = username
			if remember {
				c.Session.SetDefaultExpiration()
			} else {
				c.Session.SetNoExpiration()
			}
			c.Flash.Success("Welcome, " + username)
			return c.Redirect(routes.App.Console())
		}
	}

	c.Flash.Out["username"] = username
	c.Flash.Error("Login failed")
	return c.Redirect(routes.App.Index())
}

func (c App) Logout() revel.Result {
	for k := range c.Session {
		delete(c.Session, k)
	}
	return c.Redirect(routes.App.Index())
}

func (c App) retrieveSettingsFromDB() (s SettingsData, err error) {
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
	keys := getAllTemplateKeys(c.Params.Values)
	c.Log.Debugf("KEYS: %v", keys)
	for k, v := range keys {
		if val := c.Params.Get(k); val != "" {
			templatesRaw = append(templatesRaw, models.NotifierTemplate{
				Body: []byte(val),
				ID:   v,
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

func (c App) HasErrorsRedirect(val interface{}) (res revel.Result, ok bool) {
	if c.Validation.HasErrors() {
		// Store the validation errors in the flash context and redirect.
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(Curl.Index), true
	}
	return revel.ErrorResult{}, false
}

func (c App) Settings() revel.Result {
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

	redirectTo := c.Params.Get("redirect")
	if redirectTo == "curl" {
		return c.Redirect(Curl.Index)
	}

	return c.Redirect(App.Index)
}

func (c App) saveSettings(s SettingsData) error {
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
