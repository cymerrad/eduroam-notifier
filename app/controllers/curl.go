package controllers

import "github.com/revel/revel"

type Curl struct {
	App
}

func (c Curl) Index() revel.Result {
	return c.Render()
}
