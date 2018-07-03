package controllers

import (
	"database/sql"
	ts "eduroam-notifier/app/ts"
	"encoding/json"
	"fmt"
	"strings"

	gorp "gopkg.in/gorp.v2"

	"golang.org/x/crypto/bcrypt"

	// comment justifing import
	_ "github.com/go-sql-driver/mysql"

	"eduroam-notifier/app/models"

	"github.com/revel/revel"
)

func init() {
	revel.OnAppStart(InitDb)
	revel.OnAppStart(InitUSOSdbm)
	revel.OnAppStart(createTestUsers, 5)
	revel.OnAppStart(createTestSettings, 6)
	revel.OnAppStart(initializeGlobalVariables, 7)

	revel.InterceptMethod((*GorpController).Begin, revel.BEFORE)
	revel.InterceptMethod((*GorpController).Commit, revel.AFTER)
	revel.InterceptMethod((*GorpController).Rollback, revel.FINALLY)

	// revel.InterceptMethod(App.checkUser, revel.BEFORE)
	revel.InterceptMethod(App.AddUser, revel.BEFORE)
	revel.InterceptMethod(Curl.checkUser, revel.BEFORE)
}

func getParamString(param string, defaultValue string) string {
	p, found := revel.Config.String(param)
	if !found {
		if defaultValue == "" {
			revel.ERROR.Fatal("Cound not find parameter: " + param)
		} else {
			return defaultValue
		}
	}
	return p
}

func getParamBool(param string, defaultValue bool) bool {
	p, found := revel.Config.Bool(param)
	if !found {
		return defaultValue
	}
	return p
}

func createConnectionString(host, port, user, pass, dbname, protocol, dbargs string) string {
	if strings.Trim(dbargs, " ") != "" {
		dbargs = "?" + dbargs
	} else {
		dbargs = ""
	}

	return fmt.Sprintf("%s:%s@%s([%s]:%s)/%s%s",
		user, pass, protocol, host, port, dbname, dbargs)
}

func getConnectionString() string {
	host := getParamString("db.host", "")
	port := getParamString("db.port", "3306")
	user := getParamString("db.user", "")
	pass := getParamString("db.password", "")
	dbname := getParamString("db.name", "eduroam")
	protocol := getParamString("db.protocol", "tcp")
	dbargs := getParamString("dbargs", " ")

	return createConnectionString(host, port, user, pass, dbname, protocol, dbargs)
}

var drop bool

func conditionalDropTable(dbm *gorp.DbMap, tapmadl string) {
	if !drop {
		return
	}
	_, err := dbm.Exec("drop table " + tapmadl + " ;")
	if err != nil {
		revel.AppLog.Warnf("Error dropping '%v': %s", tapmadl, err.Error())
		return
	}
	revel.AppLog.Infof("Dropped table '%v'", tapmadl)
}

var InitDb = func() {
	connectionString := getConnectionString()
	if db, err := sql.Open("mysql", connectionString); err != nil {
		revel.ERROR.Fatal(err)
	} else {
		Dbm = &gorp.DbMap{
			Db:      db,
			Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	}
	// Defines the table for use by GORP
	drop = getParamBool("db.dropCreate", false)

	defineEventTable(Dbm)
	defineUserTable(Dbm)
	defineMessageTable(Dbm)
	defineNotifierTables(Dbm)
	defineMailMessageTable(Dbm)
	defineOptOutTable(Dbm)

	if err := Dbm.CreateTablesIfNotExists(); err != nil {
		revel.AppLog.Fatalf("Creating tables: %s", err.Error())
	}
}

func defineEventTable(dbm *gorp.DbMap) {
	conditionalDropTable(dbm, "Event")

	// set "id" as primary key and autoincrement
	t := dbm.AddTable(models.Event{}).SetKeys(true, "ID")
	t.ColMap("Body").SetNotNull(true)
}

func defineMessageTable(dbm *gorp.DbMap) {
	conditionalDropTable(dbm, "Message")

	// set "id" as primary key and autoincrement
	_ = dbm.AddTable(models.Message{}).SetKeys(true, "ID")
}

func defineUserTable(dbm *gorp.DbMap) {
	conditionalDropTable(dbm, "User")

	setColumnSizes := func(t *gorp.TableMap, colSizes map[string]int) {
		for col, size := range colSizes {
			t.ColMap(col).MaxSize = size
		}
	}
	t := dbm.AddTable(models.User{}).SetKeys(true, "ID")
	t.ColMap("Password").Transient = true
	setColumnSizes(t, map[string]int{
		"Username": 20,
		"Name":     100,
	})
}

func defineNotifierTables(dbm *gorp.DbMap) {
	conditionalDropTable(dbm, "NotifierRule")
	conditionalDropTable(dbm, "NotifierTemplate")
	conditionalDropTable(dbm, "NotifierSettings")

	t1 := dbm.AddTable(models.NotifierRule{}).SetKeys(true, "ID")
	t1.ColMap("Created").Transient = true
	t2 := dbm.AddTable(models.NotifierTemplate{}).SetKeys(true, "ID")
	t2.ColMap("Created").Transient = true
	t := dbm.AddTable(models.NotifierSettings{}).SetKeys(true, "ID")
	t.ColMap("Created").Transient = true

	t, t1 = t1, t // so the compiler won't complain
	t, t2 = t2, t
}

func defineMailMessageTable(dbm *gorp.DbMap) {
	conditionalDropTable(dbm, "MailMessage")

	// set "id" as primary key and autoincrement
	t := dbm.AddTable(models.MailMessage{}).SetKeys(true, "ID")
	t.ColMap("BodyString").Transient = true
}

func defineOptOutTable(dbm *gorp.DbMap) {
	conditionalDropTable(dbm, "OptOut")

	// set "id" as primary key and autoincrement
	_ = dbm.AddTable(models.OptOut{}).SetKeys(true, "ID")
}

func createTestUsers() {
	dUser := &models.User{}
	res, err := Dbm.Select(dUser, "Select * from User where Username='demo' limit 1;")
	if err != nil || len(res) == 0 {
		// doesn't exist -> create
		bcryptPassword, _ := bcrypt.GenerateFromPassword(
			[]byte("demo"), bcrypt.DefaultCost)
		demoUser := &models.User{0, "Demo User", "demo", "demo", bcryptPassword}
		if err := Dbm.Insert(demoUser); err != nil {
			panic(err)
		}
		revel.AppLog.Info("Created user 'demo'.")
		return
	}
	revel.AppLog.Info("User 'demo' already exists.")
}

func createTestSettings() {
	exampleSetting := ts.StartingSettings
	exampleTemplate := ts.StartingTemplate
	exampleRules := ts.StartingRules

	// insert if not existent... or not... it depends
	btz, _ := json.MarshalIndent(exampleSetting, "", "  ")
	demoSettings := &models.NotifierSettings{
		JSON:    btz,
		Created: ts.TimeZero,
	}
	tSet := &models.NotifierSettings{}
	res, err := Dbm.Select(tSet, models.GetNotifierSettings)
	if err != nil || len(res) == 0 {
		err := Dbm.Insert(demoSettings)
		if err != nil {
			revel.AppLog.Errorf("Inserting settings: %s", err.Error())
		}
		revel.AppLog.Info("Created example settings.")
	}

	tTemp := &models.NotifierTemplate{}
	res, err = Dbm.Select(tTemp, models.GetAllNotifierTemplates)
	if err != nil || len(res) == 0 {
		if err := Dbm.Insert(&exampleTemplate); err != nil {
			revel.AppLog.Errorf("Inserting template: %s", err.Error())
		}
		revel.AppLog.Info("Created example template.")
	}

	tRule := &models.NotifierRule{}
	res, err = Dbm.Select(tRule, models.GetAllNotifierRules)
	if err != nil || len(res) == 0 {
		for _, rule := range exampleRules {
			err := Dbm.Insert(&rule)
			if err != nil {
				revel.AppLog.Errorf("Inserting rule %s: %s", rule.Value, err.Error())
			}
		}

		revel.AppLog.Infof("Created %d rules", len(exampleRules))
	}

}

func initializeGlobalVariables() {
	var chillax = func(res []interface{}, err error) {
		if err != nil {
			revel.AppLog.Errorf("Failed initialization: %s", err.Error())
		}
	}
	var templates []models.NotifierTemplate
	var rules []models.NotifierRule
	var settings models.NotifierSettings
	var err error

	chillax(Dbm.Select(&templates, models.GetAllNotifierTemplates))
	chillax(Dbm.Select(&rules, models.GetAllNotifierRules))
	err = Dbm.SelectOne(&settings, models.GetNotifierSettings)
	if err != nil {
		revel.AppLog.Errorf("Failed initialization: %s", err.Error())
	}

	settingsParsed, err := settings.Unmarshall()
	if err != nil {
		revel.AppLog.Critf("Failed settings unmarshalling: %s", err.Error())
	}

	globalTemplate, err = ts.New(settingsParsed, rules, templates)
	if err != nil {
		revel.AppLog.Critf("Failed initialization of templates: %s", err.Error())
	}

	revel.AppLog.Debugf("Templating system: \n %s", globalTemplate.Show())
}

var InitUSOSdbm = func() {
	host := getParamString("usos_db.host", "")
	port := getParamString("usos_db.port", "3306")
	user := getParamString("usos_db.user", "")
	pass := getParamString("usos_db.password", "")
	dbname := getParamString("usos_db.name", "usosweb_today")
	protocol := getParamString("usos_db.protocol", "tcp")
	dbargs := getParamString("dbargs", " ")

	connectionString := createConnectionString(host, port, user, pass, dbname, protocol, dbargs)

	if db, err := sql.Open("mysql", connectionString); err != nil {
		revel.ERROR.Fatal(err)
	} else {
		USOSdbm = &gorp.DbMap{
			Db:      db,
			Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	}
}
