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
	GetAllMessagesLikeByMac = func(msg Message) string {
		sql, _, _ := sq.Select("*").From("Message").Where("Mac = '?'", msg.Mac).ToSql()
		return sql
	}
	GetAllMessagesLikeByPesel = func(msg Message) string {
		sql, _, _ := sq.Select("*").From("Message").Where("Pesel = '?'", msg.Pesel).ToSql()
		return sql
	}
	GetAllMessagesLikeByUsername = func(msg Message) string {
		sql, _, _ := sq.Select("*").From("Message").Where("Username = '?'", msg.Username).ToSql()
		return sql
	}
	GetCountMessagesLikeByMac = func(msg Message) string {
		sql, _, _ := sq.Select("COUNT(*)").From("Message").Where("Mac = '?'", msg.Mac).ToSql()
		return sql
	}
	GetCountMessagesLikeByPesel = func(msg Message) string {
		sql, _, _ := sq.Select("COUNT(*)").From("Message").Where("Pesel = '?'", msg.Pesel).ToSql()
		return sql
	}
	GetCountMessagesLikeByUsername = func(msg Message) string {
		sql, _, _ := sq.Select("COUNT(*)").From("Message").Where("Username = '?'", msg.Username).ToSql()
		return sql
	}
)
