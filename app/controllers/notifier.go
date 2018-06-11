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
	_, err := c.parseEvent()
	if err != nil {
		c.Log.Error("Error parsing event")
		c.Response.Status = http.StatusNotAcceptable
		return c.RenderText(err.Error())
	}

	event := models.Event{
		Body: c.Params.JSON,
	}

	if err := c.Txn.Insert(&event); err != nil {
		c.Log.Errorf("Error inserting event into DB: %s", err.Error())
		c.Response.Status = http.StatusNotAcceptable
		return c.RenderText(err.Error())
	}

	c.Log.Debugf("Success inserting %#v", event)

	return c.RenderText("success")
}
