package jobs

import (
	"eduroam-notifier/app/controllers"

	"github.com/revel/revel"

	"github.com/revel/modules/jobs/app/jobs"
)

// Periodically refresh USOS connection
type USOSConnection struct{}

func (c USOSConnection) Run() {
	revel.AppLog.Debugf("Refreshing USOS connection")
	controllers.InitUSOSdbm()
}

func init() {
	revel.OnAppStart(func() {
		jobs.Schedule("@every 1h", USOSConnection{})
	})
}
