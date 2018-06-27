package models

import (
	sq "gopkg.in/Masterminds/squirrel.v1"
)

const (
	GetAllNotifierRules     = "SELECT * FROM NotifierRule WHERE CreatedString = ( SELECT MAX(CreatedString) FROM NotifierRule )"
	GetAllNotifierTemplates = "SELECT * FROM NotifierTemplate WHERE CreatedString = ( SELECT MAX(CreatedString) FROM NotifierTemplate )"
	GetNotifierSettings     = "SELECT * FROM NotifierSettings WHERE CreatedString = ( SELECT MAX(CreatedString) FROM NotifierSettings ) LIMIT 1"
)

var (
	// I don't like constant strings in code - they are hard to edit later on
	partialGetCount            = func() sq.SelectBuilder { return sq.Select("COUNT(*)") }
	partialGetAll              = func() sq.SelectBuilder { return sq.Select("*") }
	partialGetCountFromMessage = func() sq.SelectBuilder { return partialGetCount().From("Message") }
	partialGetAllFromMessage   = func() sq.SelectBuilder { return partialGetAll().From("Message") }

	GetAllMessagesLikeByMac = func(msg Message) string {
		sql, _, _ := partialGetAllFromMessage().Where("Mac = '?'", msg.Mac).ToSql()
		return sql
	}
	GetAllMessagesLikeByPesel = func(msg Message) string {
		sql, _, _ := partialGetAllFromMessage().Where("Pesel = '?'", msg.Pesel).ToSql()
		return sql
	}
	GetAllMessagesLikeByUsername = func(msg Message) string {
		sql, _, _ := partialGetAllFromMessage().Where("Username = '?'", msg.Username).ToSql()
		return sql
	}
	GetCountMessagesLikeByMac = func(msg Message) string {
		sql, _, _ := partialGetCountFromMessage().Where("Mac = '?'", msg.Mac).ToSql()
		return sql
	}
	GetCountMessagesLikeByPesel = func(msg Message) string {
		sql, _, _ := partialGetCountFromMessage().Where("Pesel = '?'", msg.Pesel).ToSql()
		return sql
	}
	GetCountMessagesLikeByUsername = func(msg Message) string {
		sql, _, _ := partialGetCountFromMessage().Where("Username = '?'", msg.Username).ToSql()
		return sql
	}

	GetOptOutsOfUser = func(msg Message) string {
		sql, _, _ := partialGetAll().From("OptOut").Where("Username = '?'", msg.Username).ToSql()
		return sql
	}
)
