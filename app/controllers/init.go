package controllers

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/go-gorp/gorp"
	"golang.org/x/crypto/bcrypt"

	// comment justifing import
	_ "github.com/go-sql-driver/mysql"

	"eduroam-notifier/app/models"

	"github.com/revel/revel"
)

func init() {
	revel.OnAppStart(InitDb)
	revel.OnAppStart(createTestUsers, 5)

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

func getConnectionString() string {
	host := getParamString("db.host", "")
	port := getParamString("db.port", "3306")
	user := getParamString("db.user", "")
	pass := getParamString("db.password", "")
	dbname := getParamString("db.name", "auction")
	protocol := getParamString("db.protocol", "tcp")
	dbargs := getParamString("dbargs", " ")

	if strings.Trim(dbargs, " ") != "" {
		dbargs = "?" + dbargs
	} else {
		dbargs = ""
	}
	return fmt.Sprintf("%s:%s@%s([%s]:%s)/%s%s",
		user, pass, protocol, host, port, dbname, dbargs)
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
	_ = dbm.AddTable(models.Message{}).SetKeys(false, "ID")
}

func defineUserTable(dbm *gorp.DbMap) {
	conditionalDropTable(dbm, "User")

	setColumnSizes := func(t *gorp.TableMap, colSizes map[string]int) {
		for col, size := range colSizes {
			t.ColMap(col).MaxSize = size
		}
	}
	t := dbm.AddTable(models.User{}).SetKeys(true, "UserId")
	t.ColMap("Password").Transient = true
	setColumnSizes(t, map[string]int{
		"Username": 20,
		"Name":     100,
	})
}

func defineNotifierTables(dbm *gorp.DbMap) {
	conditionalDropTable(dbm, "NotifierRule")
	conditionalDropTable(dbm, "NotifierSettings")

	t1 := dbm.AddTable(models.NotifierRule{}).SetKeys(true, "ID")
	t := dbm.AddTable(models.NotifierSettings{})

	t, t1 = t1, t // so the compiler won't complain
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
