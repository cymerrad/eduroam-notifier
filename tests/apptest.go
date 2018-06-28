package tests

import (
	"bytes"
	"encoding/json"

	"github.com/revel/revel/testing"
)

type AppTest struct {
	testing.TestSuite
}

func (t *AppTest) Before() {
	println("Set up")
}

func (t *AppTest) TestThatIndexPageWorks() {
	t.Get("/")
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")
}

func (t *AppTest) TestThatCurlPageWorks() {
	t.Get("/curl")
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")
}

func (t *AppTest) PostingToRoot() {
	btz, err := json.Marshal(ExampleInput)
	t.AssertEqual(nil, err)
	rdr := bytes.NewReader(btz)
	t.Post("/", "application/json", rdr)
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")
}

func (t *AppTest) After() {
	println("Tear down")
}
