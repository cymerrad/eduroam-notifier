package controllers

import (
	"eduroam-notifier/app/models"
	"eduroam-notifier/app/template_system"
	"errors"
	"net/http"
	"strconv"
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

		interpretMessage(match.Fields, nil, globalTemplate)

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

func (c App) interpretEvent(event models.EventParsed, templateSystem *template_system.T) ([]models.MailMessage, error) {
	out := make([]models.MailMessage, 0)

	for _, match := range event.CheckResult.MatchingMessages {
		extras := make(map[string]string)
		msg := match.ToMessage(0)

		// PRE-RENDER CHECKS

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

		result := interpretMessage(match.Fields, extras, templateSystem)
		out = append(out, result)
	}

	return out, nil
}

func interpretMessage(fields models.EventMessageFields, extras map[string]string, a *template_system.T) (resp models.MailMessage) {
	revel.AppLog.Debugf("Doing something magical with %#v", fields)

	output, err := a.Input(fields, extras)
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

// TODO
// this might seriously change (e.g. calls to some other service)
func determineRecipient(fields models.EventMessageFields) (string, error) {
	if fields.SourceUser == "" {
		return "", errors.New("empty SourceUser field")
	}
	return fields.SourceUser, nil
}
