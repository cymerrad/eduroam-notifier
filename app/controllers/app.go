package controllers

import (
	"database/sql"
	"eduroam-notifier/app/models"
	"eduroam-notifier/app/routes"
	"fmt"

	"github.com/revel/modules/orm/gorp/app/controllers"
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	gorpController.Controller
}

func (c App) Index() revel.Result {
	return c.Render()
}

func (c App) Console() revel.Result {
	return c.Render()
}

func (c App) Notify() revel.Result {
	return c.RenderJSON(struct{ Message, Error string }{"lol", "hej"})
}

func (c App) getUser(username string) (user *models.User) {
	user = &models.User{}
	fmt.Println("get user", username, c.Txn)

	err := c.Txn.SelectOne(user, c.Db.SqlStatementBuilder.Select("*").From("User").Where("Username=?", username))
	if err != nil {
		if err != sql.ErrNoRows {
			c.Log.Error("Failed to find user")
		}
		return nil
	}
	return
}

func (c App) Login(username, password string, remember bool) revel.Result {
	user := c.getUser(username)
	if user != nil {
		err := bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password))
		if err == nil {
			c.Session["user"] = username
			if remember {
				c.Session.SetDefaultExpiration()
			} else {
				c.Session.SetNoExpiration()
			}
			c.Flash.Success("Welcome, " + username)
			return c.Redirect(routes.App.Console())
		}
	}

	c.Flash.Out["username"] = username
	c.Flash.Error("Login failed")
	return c.Redirect(routes.App.Index())
}

func (c App) Logout() revel.Result {
	for k := range c.Session {
		delete(c.Session, k)
	}
	return c.Redirect(routes.App.Index())
}
