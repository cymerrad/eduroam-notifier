package models

const (
	GetAllNotifierRules     = "SELECT * FROM NotifierRule WHERE CreatedString = ( SELECT MAX(CreatedString) FROM NotifierRule )"
	GetAllNotifierTemplates = "SELECT * FROM NotifierTemplate WHERE CreatedString = ( SELECT MAX(CreatedString) FROM NotifierTemplate )"
	GetNotifierSettings     = "SELECT * FROM NotifierSettings WHERE CreatedString = ( SELECT MAX(CreatedString) FROM NotifierSettings ) LIMIT 1"
)
