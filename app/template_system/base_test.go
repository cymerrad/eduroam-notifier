package ts

import (
	"eduroam-notifier/app/models"
	"fmt"
	"reflect"
	"testing"
	"time"
)

var timeZero = time.Unix(0, 0)

func TestParseRules(t *testing.T) {
	type args struct {
		rules []models.NotifierRule
	}
	tests := []struct {
		name     string
		args     args
		wantOutA map[Action]TemplateID
		wantOutF map[TemplateTag]Field
		wantOutC map[TemplateTag]ConstValue
		wantOutI map[Action]int
		wantErr  bool
	}{
		{"test", args{[]models.NotifierRule{
			{
				On:      "template_tag",
				Do:      "insert_text",
				Value:   "{\"template_tag\" : \"signature\", \"insert_text\" : \"DSK UW\"}",
				Created: timeZero,
			},
			{
				On:      "template_tag",
				Do:      "substitute_with_field",
				Value:   "{\"template_tag\" : \"mac\", \"substitute_with_field\" : \"source-mac\"}",
				Created: timeZero,
			},
			{
				On:      "template_tag",
				Do:      "substitute_with_field",
				Value:   "{\"template_tag\" : \"pesel\", \"substitute_with_field\" : \"Pesel\"}",
				Created: timeZero,
			},
			{
				On:      OnAction,
				Do:      DoActionPickTemplate,
				Value:   fmt.Sprintf("{\"%s\" : \"Login incorrect (mschap: MS-CHAP2-Response is incorrect)\", \"%s\" : \"wrong_password\"}", OnAction, DoActionPickTemplate),
				Created: timeZero,
			},
		}},
			map[Action]TemplateID{},
			map[TemplateTag]Field{},
			map[TemplateTag]ConstValue{},
			map[Action]int{},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOutA, gotOutF, gotOutC, gotOutI, err := ParseRules(tt.args.rules)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOutA, tt.wantOutA) {
				t.Errorf("ParseRules() gotOutA = %v, want %v", gotOutA, tt.wantOutA)
			}
			if !reflect.DeepEqual(gotOutF, tt.wantOutF) {
				t.Errorf("ParseRules() gotOutF = %v, want %v", gotOutF, tt.wantOutF)
			}
			if !reflect.DeepEqual(gotOutC, tt.wantOutC) {
				t.Errorf("ParseRules() gotOutC = %v, want %v", gotOutC, tt.wantOutC)
			}
			if !reflect.DeepEqual(gotOutI, tt.wantOutI) {
				t.Errorf("ParseRules() gotOutI = %v, want %v", gotOutI, tt.wantOutI)
			}
		})
	}
}
