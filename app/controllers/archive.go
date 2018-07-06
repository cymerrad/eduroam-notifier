package controllers

import "github.com/revel/revel"

type Archive struct {
	App
}

func (c Archive) Index() revel.Result {
	if c.Validation.HasErrors() {
		return c.Render()
	}

	c.ViewArgs["archive"] = map[string]string{"hello": "there"}

	return c.Render()
}
