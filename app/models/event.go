package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/revel/revel"
	gorp "gopkg.in/gorp.v2"
)

type Event struct {
	ID            int
	Body          []byte
	CreatedString string

	// transient
	Created time.Time
}

func (o *Event) PreInsert(_ gorp.SqlExecutor) error {
	o.CreatedString = o.Created.Format(time.RFC3339)

	return nil
}

func (o *Event) PostGet(_ gorp.SqlExecutor) error {
	t, err := time.Parse(time.RFC3339, o.CreatedString)
	if err != nil {
		return err
	}
	o.Created = t
	return nil
}

// Event received (autogenerated)
type EventParsed struct {
	CheckResult struct {
		ResultDescription  string `json:"result_description"`
		TriggeredCondition struct {
			ID            string    `json:"id"`
			Type          string    `json:"type"`
			CreatedAt     time.Time `json:"created_at"`
			CreatorUserID string    `json:"creator_user_id"`
			Title         string    `json:"title"`
			Parameters    struct {
				Grace               int    `json:"grace"`
				Backlog             int    `json:"backlog"`
				RepeatNotifications bool   `json:"repeat_notifications"`
				Field               string `json:"field"`
				Value               string `json:"value"`
			} `json:"parameters"`
		} `json:"triggered_condition"`
		TriggeredAt      time.Time              `json:"triggered_at"`
		Triggered        bool                   `json:"triggered"`
		MatchingMessages []EventMatchingMessage `json:"matching_messages"`
	} `json:"check_result"`
	Stream struct {
		CreatorUserID string        `json:"creator_user_id"`
		Outputs       []interface{} `json:"outputs"`
		Description   string        `json:"description"`
		CreatedAt     time.Time     `json:"created_at"`
		Rules         []struct {
			Field       string `json:"field"`
			StreamID    string `json:"stream_id"`
			Description string `json:"description"`
			ID          string `json:"id"`
			Type        int    `json:"type"`
			Inverted    bool   `json:"inverted"`
			Value       string `json:"value"`
		} `json:"rules"`
		AlertConditions []struct {
			CreatorUserID string    `json:"creator_user_id"`
			CreatedAt     time.Time `json:"created_at"`
			ID            string    `json:"id"`
			Type          string    `json:"type"`
			Title         string    `json:"title"`
			Parameters    struct {
				Grace               int    `json:"grace"`
				Backlog             int    `json:"backlog"`
				RepeatNotifications bool   `json:"repeat_notifications"`
				Field               string `json:"field"`
				Value               string `json:"value"`
			} `json:"parameters"`
		} `json:"alert_conditions"`
		Title                          string      `json:"title"`
		ContentPack                    interface{} `json:"content_pack"`
		IsDefaultStream                bool        `json:"is_default_stream"`
		IndexSetID                     string      `json:"index_set_id"`
		MatchingType                   string      `json:"matching_type"`
		RemoveMatchesFromDefaultStream bool        `json:"remove_matches_from_default_stream"`
		Disabled                       bool        `json:"disabled"`
		ID                             string      `json:"id"`
	} `json:"stream"`
}

type EventMatchingMessage struct {
	Index     string             `json:"index"`
	Message   string             `json:"m"`
	Fields    EventMessageFields `json:"fields"`
	ID        string             `json:"id"`
	Timestamp time.Time          `json:"timestamp"`
	Source    string             `json:"source"`
	StreamIds []string           `json:"stream_ids"`
}

// TODO
// this probably should be map[string]string or map[string]interface{}
// I need to marshall this and unmarshall it back a few times for using it in the template
type EventMessageFields struct {
	Level          int    `json:"-"` // in fact it is a "level"
	Gl2RemoteIP    string `json:"gl2_remote_ip"`
	Gl2RemotePort  int    `json:"-"` // and this is "gl2_remote_port", but don't tell anyone that I ignored these fields
	SourceUser     string `json:"source-user"`
	Gl2SourceInput string `json:"gl2_source_input"`
	EDUROAMACT     string `json:"EDUROAM_ACT"`
	WINDOWSMAC     string `json:"WINDOWSMAC"`
	SourceMac      string `json:"source-mac"`
	Pesel          string `json:"Pesel"`
	Username       string `json:"Username"`
	USERNAME       string `json:"USERNAME"`
	Action         string `json:"action"`
	Client         string `json:"client"`
	Gl2SourceNode  string `json:"gl2_source_node"`
	Facility       string `json:"facility"`
	Realm          string `json:"Realm"`
}

func (u *Event) String() string {
	return fmt.Sprintf("Event(ID: %d)", u.ID)
}

func (u *Event) Parse() (EventParsed, error) {
	out := EventParsed{}
	err := json.Unmarshal(u.Body, &out)
	return out, err
}

const mediumblobMaxSize = 16777215

func (u *Event) Validate(v *revel.Validation) {
	v.Required(u.Body)
	v.MaxSize(u.Body, mediumblobMaxSize)
	if _, err := u.Parse(); err != nil {
		v.ValidationResult(false).Message(err.Error())
		return
	}
	v.ValidationResult(true)
}

func (m *EventMatchingMessage) ToMessage(eventID int) Message {
	return Message{
		ID:        m.ID,
		EventID:   eventID,
		Message:   m.Message,
		Timestamp: m.Timestamp,
		Pesel:     m.Fields.Pesel,
		Username:  m.Fields.SourceUser,
		Mac:       m.Fields.SourceMac,
		Action:    m.Fields.Action,
		Realm:     m.Fields.Realm,
		Facility:  m.Fields.Facility,
	}
}
