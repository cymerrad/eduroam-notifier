package controllers

import (
	"database/sql"
	"eduroam-notifier/app/models"
	"eduroam-notifier/app/routes"

	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"

	sq "gopkg.in/Masterminds/squirrel.v1"
)

type App struct {
	GorpController
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
	c.Log.Debugf("Get user %s %v", username, c.Txn)

	str, _, err := sq.StatementBuilder.Select("*").From("User").Where(sq.Eq{"Username": username}).ToSql()
	if err != nil {
		c.Log.Errorf("Failed to build query")
		return nil
	}
	err = c.Txn.SelectOne(user, str, username) // why do I have to pass the 'username' second time?
	if err != nil {
		c.Log.Debugf("Failed query: %s; (%v)", str, user)
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
