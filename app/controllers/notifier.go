package controllers

import (
	"eduroam-notifier/app/models"
	"eduroam-notifier/app/template_system"
	"errors"
	"net/http"
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

		interpretMessage(match.Fields, globalTemplate)

	}

	return c.RenderText("k")
}

func (c Notifier) parseEvent() (models.EventParsed, error) {
	eventP := models.EventParsed{}
	err := c.Params.BindJSON(&eventP)
	return eventP, err
}

// send mail
type ResponseAction struct {
	Recipient string `json:"recipient"`
	Body      string `json:"body"`
	Error     string `json:"error,omitempty"`
}

func interpretMessage(fields models.EventMessageFields, a *template_system.T) (resp ResponseAction) {
	revel.AppLog.Debugf("Doing something magical with %#v", fields)

	output, err := a.Input(fields)
	if err != nil {
		resp.Error = err.Error()
		resp.Recipient = "none (do nothing)"
		return
	}
	resp.Body = output

	recipient, err := determineRecipient(fields)
	if err != nil {
		resp.Error = err.Error()
		resp.Recipient = "(cannot be found)"
	}
	resp.Recipient = recipient

	return
}

// this might seriously change (e.g. calls to some other service)
func determineRecipient(fields models.EventMessageFields) (string, error) {
	if fields.SourceUser == "" {
		return "", errors.New("empty SourceUser field")
	}
	return fields.SourceUser, nil
}
