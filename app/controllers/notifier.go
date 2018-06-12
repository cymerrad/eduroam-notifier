package controllers

import (
	"eduroam-notifier/app/models"
	"net/http"

	"github.com/revel/revel"
)

type Notifier struct {
	App
}

var settings models.NotifierSettings

func (c Notifier) Notify() revel.Result {
	parsedEvent, err := c.parseEvent()
	if err != nil {
		c.Log.Error("Error parsing event")
		c.Response.Status = http.StatusNotAcceptable
		return c.RenderText(err.Error())
	}

	for _, message := range parsedEvent.CheckResult.MatchingMessages {
		sourceUser := message.Fields.SourceUser
		sourceMac := message.Fields.SourceMac
		timestamp := message.Timestamp

		event := models.Event{
			Body:      c.Params.JSON,
			Mac:       sourceMac,
			Username:  sourceUser,
			Timestamp: timestamp,
		}

		if err := c.Txn.Insert(&event); err != nil {
			c.Log.Errorf("Error inserting event into DB: %s", err.Error())
			continue
		}

		c.Log.Debugf("Success inserting event ID %d", event.ID)

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

func interpretMessage(msg models.EventMatchingMessage) ResponseAction {
	sourceUser := msg.Fields.SourceUser
	sourceMac := msg.Fields.SourceMac
	timestamp := msg.Timestamp

}
