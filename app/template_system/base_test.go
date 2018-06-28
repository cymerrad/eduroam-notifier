package ts

import (
	"eduroam-notifier/app/models"
	"reflect"
	"testing"
	"time"
)

var TimeZero = time.Unix(0, 0)
var GoodStartingSettings = []models.NotifierRule{
	{
		On:      OnTemplateTag,
		Do:      DoInsertText,
		Value:   GenerateJSON(OnTemplateTag, "signature", DoInsertText, "DSK UW"),
		Created: TimeZero,
	},
	{
		On:      OnTemplateTag,
		Do:      DoSubstituteWithField,
		Value:   GenerateJSON(OnTemplateTag, "mac", DoSubstituteWithField, "source-mac"),
		Created: TimeZero,
	},
	{
		On:      OnTemplateTag,
		Do:      DoSubstituteWithField,
		Value:   GenerateJSON(OnTemplateTag, "pesel", DoSubstituteWithField, "Pesel"),
		Created: TimeZero,
	},
	{
		On:      OnAction,
		Do:      DoActionPickTemplate,
		Value:   GenerateJSON(OnAction, "Login incorrect (mschap: MS-CHAP2-Response is incorrect)", DoActionPickTemplate, "wrong_password"),
		Created: TimeZero,
	},
	{
		On:      OnAction,
		Do:      DoIgnoreFirstN,
		Value:   GenerateJSON(OnAction, "Login incorrect (mschap: MS-CHAP2-Response is incorrect)", DoIgnoreFirstN, "5"),
		Created: TimeZero,
	},
}

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
		{"test", args{GoodStartingSettings},
			map[Action]TemplateID{"Login incorrect (mschap: MS-CHAP2-Response is incorrect)": "wrong_password"},
			map[TemplateTag]Field{"mac": "source-mac", "pesel": "Pesel"},
			map[TemplateTag]ConstValue{"signature": "DSK UW"},
			map[Action]int{"Login incorrect (mschap: MS-CHAP2-Response is incorrect)": 5},
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
