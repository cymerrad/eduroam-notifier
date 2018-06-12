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

	for _, message := range parsedEvent.CheckResult.MatchingMessages {
		msg := models.Message{
			EventMessageFields: message.Fields,
			ID:                 message.ID,
			Message:            message.Message,
			Timestamp:          message.Timestamp,
		}

		if err := c.Txn.Insert(&msg); err != nil {
			c.Log.Errorf("Error inserting event into DB: %s", err.Error())
			continue
		}

		c.Log.Debugf("Success inserting message %#v", msg)

	}

	return c.RenderText("k")
}

func (c Notifier) parseEvent() (models.EventParsed, error) {
	eventP := models.EventParsed{}
	err := c.Params.BindJSON(&eventP)
	return eventP, err
}

type ResponseAction struct {
}

func interpretMessage(msg models.Message, settings models.NotifierSettings) ResponseAction {
	revel.AppLog.Debugf("Doing something magical with %#v", msg)

	return ResponseAction{}
}
