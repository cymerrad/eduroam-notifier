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
	partialGetCount            = func() sq.SelectBuilder { return mysql.Select("COUNT(*)") }
	partialGetAll              = func() sq.SelectBuilder { return sq.Select("*") }
	partialGetCountFromMessage = func() sq.SelectBuilder { return partialGetCount().From("Message") }
	partialGetAllFromMessage   = func() sq.SelectBuilder { return partialGetAll().From("Message") }

	GetAllMessagesLikeByMac = func(msg Message) string {
		sql, _, _ := partialGetAllFromMessage().Where(thisLibrarySucks("Mac", msg.Mac)).ToSql()
		return sql
	}
	GetAllMessagesLikeByPesel = func(msg Message) string {
		sql, _, _ := partialGetAllFromMessage().Where(thisLibrarySucks("Pesel", msg.Pesel)).ToSql()
		return sql
	}
	GetAllMessagesLikeByUsername = func(msg Message) string {
		sql, _, _ := partialGetAllFromMessage().Where(thisLibrarySucks("Username", msg.Username)).ToSql()
		return sql
	}
	GetCountMessagesLikeByMac = func(msg Message) string {
		sql, _, _ := partialGetCountFromMessage().Where(thisLibrarySucks("Mac", msg.Mac)).ToSql()
		return sql
	}
	GetCountMessagesLikeByPesel = func(msg Message) string {
		sql, _, _ := partialGetCountFromMessage().Where(thisLibrarySucks("Pesel", msg.Pesel)).ToSql()
		return sql
	}
	GetCountMessagesLikeByUsername = func(msg Message) string {
		sql, _, _ := partialGetCountFromMessage().Where(thisLibrarySucks("Username", msg.Username)).ToSql()
		return sql
	}

	GetOptOutsOfUser = func(msg Message) string {
		sql, _, _ := partialGetAll().From("OptOut").Where(thisLibrarySucks("Username", msg.Username)).ToSql()
		return sql
	}
)
