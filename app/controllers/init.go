package controllers

import (
	"github.com/revel/revel"
)

func init() {

	// revel.InterceptMethod(App.checkUser, revel.BEFORE)
	revel.InterceptMethod(App.AddUser, revel.BEFORE)
	revel.InterceptMethod(Curl.checkUser, revel.BEFORE)
}
