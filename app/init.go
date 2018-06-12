package app

import (
	"eduroam-notifier/app/models"
	// comment justifing this
	_ "github.com/go-sql-driver/mysql"
	rgorp "github.com/revel/modules/orm/gorp/app"
	"golang.org/x/crypto/bcrypt"
	gorp "gopkg.in/gorp.v2"

	"github.com/revel/revel"
)

var (
	// AppVersion revel app version (ldflags)
	AppVersion string

	// BuildTime revel app build-time (ldflags)
	BuildTime string
)

func init() {
	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		revel.SessionFilter,           // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		revel.I18nFilter,              // Resolve the requested language
		HeaderFilter,                  // Add some security based headers
		revel.InterceptorFilter,       // Run interceptors around the action.
		revel.CompressFilter,          // Compress the result.
		revel.ActionInvoker,           // Invoke the action.
	}

	// Register startup functions with OnAppStart
	// revel.DevMode and revel.RunMode only work inside of OnAppStart. See Example Startup Script
	// ( order dependent )
	// revel.OnAppStart(ExampleStartupScript)
	// revel.OnAppStart(InitDB)
	// revel.OnAppStart(FillCache)

	revel.OnAppStart(func() {
		drop = getParamBool("db.dropCreate", false)

		Dbm := rgorp.Db.Map
		defineEventTable(Dbm)
		defineUserTable(Dbm)
		defineNotifierTables(Dbm)

		rgorp.Db.TraceOn(revel.AppLog)
		Dbm.CreateTablesIfNotExists()

		createTestUsers(Dbm)
	}, 5)

}

// HeaderFilter adds common security headers
// There is a full implementation of a CSRF filter in
// https://github.com/revel/modules/tree/master/csrf
var HeaderFilter = func(c *revel.Controller, fc []revel.Filter) {
	c.Response.Out.Header().Add("X-Frame-Options", "SAMEORIGIN")
	c.Response.Out.Header().Add("X-XSS-Protection", "1; mode=block")
	c.Response.Out.Header().Add("X-Content-Type-Options", "nosniff")
	c.Response.Out.Header().Add("Referrer-Policy", "strict-origin-when-cross-origin")

	fc[0](c, fc[1:]) // Execute the next filter stage.
}

//func ExampleStartupScript() {
//	// revel.DevMod and revel.RunMode work here
//	// Use this script to check for dev mode and set dev/prod startup scripts here!
//	if revel.DevMode == true {
//		// Dev mode
//	}
//}

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

func defineEventTable(dbm *gorp.DbMap) {
	conditionalDropTable(dbm, "Event")

	// set "id" as primary key and autoincrement
	t := dbm.AddTable(models.Event{}).SetKeys(true, "ID")
	t.ColMap("Body").SetNotNull(true)
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

	t1 := dbm.AddTable(models.NotifierRule{})
	t := dbm.AddTable(models.NotifierSettings{})

	revel.AppLog.Debugf("Experimenting with %v %v", t1, t)
}

func createTestUsers(dbm *gorp.DbMap) {
	dUser := &models.User{}
	res, err := dbm.Select(dUser, "Select * from User where Username='demo' limit 1;")
	if err != nil || len(res) == 0 {
		// doesn't exist -> create
		bcryptPassword, _ := bcrypt.GenerateFromPassword(
			[]byte("demo"), bcrypt.DefaultCost)
		demoUser := &models.User{0, "Demo User", "demo", "demo", bcryptPassword}
		if err := dbm.Insert(demoUser); err != nil {
			panic(err)
		}
		revel.AppLog.Info("Created user 'demo'.")
		return
	}
	revel.AppLog.Info("User 'demo' already exists.")
}

func getParamBool(param string, defaultValue bool) bool {
	p, found := revel.Config.Bool(param)
	if !found {
		return defaultValue
	}
	return p
}
