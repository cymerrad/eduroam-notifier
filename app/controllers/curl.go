package controllers

import "github.com/revel/revel"

type Curl struct {
	App
}

type curlData struct {
	Input, Output string
}

func (c Curl) Index() revel.Result {
	return c.Render()
}

func (c Curl) Notify() revel.Result {
	c.ViewArgs["curl"] = curlData{
		Input:  "{'lol':'powo'\n'dupa':'cycki'}",
		Output: "Previous input was " + c.Params.Get("json"),
	}

	return c.RenderTemplate("Curl/Index.html")
}
