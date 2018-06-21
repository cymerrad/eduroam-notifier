package template_system

var Schema map[string][]string = map[string][]string{
	"action":       {"send_template"},
	"template_tag": {"substitute_with_field", "insert_text"},
}
