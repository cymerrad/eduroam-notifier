package models

import (
	"fmt"

	sq "gopkg.in/Masterminds/squirrel.v1"
)

const (
	GetAllNotifierRules     = "SELECT * FROM NotifierRule WHERE CreatedString = ( SELECT MAX(CreatedString) FROM NotifierRule )"
	GetAllNotifierTemplates = "SELECT * FROM NotifierTemplate WHERE CreatedString = ( SELECT MAX(CreatedString) FROM NotifierTemplate )"
	GetNotifierSettings     = "SELECT * FROM NotifierSettings WHERE CreatedString = ( SELECT MAX(CreatedString) FROM NotifierSettings ) LIMIT 1"
)

var (
	mysql            = sq.StatementBuilder.PlaceholderFormat(sq.Question)
	thisLibrarySucks = func(column, value string) string { return fmt.Sprintf("%s = '%s'", column, value) }
	// I don't like constant strings in code - they are hard to edit later on
	partialGetCount             = func() sq.SelectBuilder { return mysql.Select("COUNT(*)") }
	partialGetAll               = func() sq.SelectBuilder { return sq.Select("*") }
	partialGetCountFromIncident = func() sq.SelectBuilder { return partialGetCount().From("Incident") }
	partialGetAllFromIncident   = func() sq.SelectBuilder { return partialGetAll().From("Incident") }
	partialGetAllOptOuts        = func() sq.SelectBuilder { return mysql.Select("ID, Mac, Pesel, Username, Action, Comment") }

	GetAllIncidentsLikeByMac = func(incid Incident) string {
		sql, _, _ := partialGetAllFromIncident().Where(thisLibrarySucks("Mac", incid.Mac)).ToSql()
		return sql
	}
	GetAllIncidentsLikeByPesel = func(incid Incident) string {
		sql, _, _ := partialGetAllFromIncident().Where(thisLibrarySucks("Pesel", incid.Pesel)).ToSql()
		return sql
	}
	GetAllIncidentsLikeByUsername = func(incid Incident) string {
		sql, _, _ := partialGetAllFromIncident().Where(thisLibrarySucks("Username", incid.Username)).ToSql()
		return sql
	}
	GetCountIncidentsLikeByMac = func(incid Incident) string {
		sql, _, _ := partialGetCountFromIncident().Where(thisLibrarySucks("Mac", incid.Mac)).ToSql()
		return sql
	}
	GetCountIncidentsLikeByPesel = func(incid Incident) string {
		sql, _, _ := partialGetCountFromIncident().Where(thisLibrarySucks("Pesel", incid.Pesel)).ToSql()
		return sql
	}
	GetCountIncidentsLikeByUsername = func(incid Incident) string {
		sql, _, _ := partialGetCountFromIncident().Where(thisLibrarySucks("Username", incid.Username)).ToSql()
		return sql
	}

	GetOptOutsOfUser = func(incid Incident) string {
		sql, _, _ := partialGetAllOptOuts().From("OptOut").Where(thisLibrarySucks("Pesel", incid.Pesel)).ToSql()
		return sql
	}
	GetLastIncidentByHash = func(hash string) string {
		sql := fmt.Sprintf("SELECT ID, EventID, Incident, Mac, Pesel, Username, Action FROM Incident WHERE Pesel IN (SELECT DISTINCT Pesel FROM MailMessage WHERE Hash='%s') ORDER BY -Timestamp LIMIT 1", hash)
		return sql
	}
)
