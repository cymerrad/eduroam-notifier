package controllers

import (
	"eduroam-notifier/app/models"
	"net/http"
	"time"

	"github.com/revel/revel"
)

type Notifier struct {
	App
}

var settings models.NotifierSettings
var rules []models.NotifierRule
var templates []models.NotifierTemplate

func (c Notifier) Notify() revel.Result {
	now := time.Now()

	parsedEvent, err := c.parseEvent()
	if err != nil {
		c.Log.Error("Error parsing event")
		c.Response.Status = http.StatusNotAcceptable
		return c.RenderText(err.Error())
	}

	event := models.Event{
		Body:      c.Params.JSON,
		Timestamp: now,
	}
	if err := c.Txn.Insert(&event); err != nil {
		c.Log.Errorf("Error inserting event into DB: %s", err.Error())
		c.Response.Status = http.StatusInternalServerError
		return c.RenderText(":c")
	}

	for _, match := range parsedEvent.CheckResult.MatchingMessages {
		msg := match.ToMessage(event.ID)

		if err := c.Txn.Insert(&msg); err != nil {
			c.Log.Errorf("Error inserting event into DB: %s", err.Error())
			continue
		}

		c.Log.Debugf("Success inserting message %#v", msg, settings)

		interpretMessage(msg, settings, rules, templates)

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
	Recipient, Body string
}

func interpretMessage(msg models.Message, stg models.NotifierSettings, rules []models.NotifierRule, templates []models.NotifierTemplate) ResponseAction {
	revel.AppLog.Debugf("Doing something magical with %#v under settings %#v", msg, stg)

	return ResponseAction{}
}
