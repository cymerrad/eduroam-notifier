package ts

import (
	"eduroam-notifier/app/models"
	"reflect"
	"testing"
	"text/template"
)

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
		wantOutS map[Action]string
		wantErr  bool
	}{
		{"valid", args{StartingRules},
			map[Action]TemplateID{"Login incorrect (mschap: MS-CHAP2-Response is incorrect)": "wrong_password"},
			map[TemplateTag]Field{"mac": "source-mac", "pesel": "Pesel"},
			map[TemplateTag]ConstValue{"signature": "DSK UW"},
			map[Action]int{"Login incorrect (mschap: MS-CHAP2-Response is incorrect)": 5},
			map[Action]string{"Login incorrect (mschap: MS-CHAP2-Response is incorrect)": "Ostrzeżenie Eduroam"},
			false},
		{ErrDeclaredValueMismatch.Error(), args{[]models.NotifierRule{
			{
				On:    OnTemplateTag,
				Do:    DoInsertText,
				Value: GenerateJSON("cokolwiek", "signature", "cokolwiek", "DSK UW"),
			},
		}},
			nil, nil, nil, nil, nil,
			true},
		{ErrUnrecognizedOption("nie").Error(), args{[]models.NotifierRule{
			{
				On:    "nie",
				Do:    "eh",
				Value: GenerateJSON(OnTemplateTag, "signature", DoInsertText, "DSK UW"),
			},
		}},
			nil, nil, nil, nil, nil,
			true},
		{ErrUnrecognizedOption("eh").Error(), args{[]models.NotifierRule{
			{
				On:    OnTemplateTag,
				Do:    "eh",
				Value: GenerateJSON(OnTemplateTag, "signature", DoInsertText, "DSK UW"),
			},
		}},
			nil, nil, nil, nil, nil,
			true},
		{"strconv.Atoi: parsing \"abc\": invalid syntax", args{[]models.NotifierRule{
			{
				On:    OnAction,
				Do:    DoActionIgnoreFirstN,
				Value: GenerateJSON(OnAction, "Login incorrect (mschap: MS-CHAP2-Response is incorrect)", DoActionIgnoreFirstN, "abc"),
			},
		}},
			nil, nil, nil, nil, nil,
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOutA, gotOutF, gotOutC, gotOutI, gotOutS, err := ParseRules(tt.args.rules)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !reflect.DeepEqual(err.Error(), tt.name) {
					t.Errorf("ParseRules() err = `%v`, want `%v`", err.Error(), tt.name)
				}
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
			if !reflect.DeepEqual(gotOutS, tt.wantOutS) {
				t.Errorf("ParseRules() gotOutS = %v, want %v", gotOutS, tt.wantOutS)
			}
		})
	}
}

func TestT_Input(t *testing.T) {
	// parsed example rules which should work, 'cause they are tested
	ts, _ := New(StartingSettings, StartingRules, []models.NotifierTemplate{StartingTemplate})

	type fields struct {
		Templates        map[TemplateID]*template.Template
		Actions          map[Action]TemplateID
		ReplaceWithField map[TemplateTag]Field
		ReplaceWithConst map[TemplateTag]ConstValue
		IgnoreFirst      map[Action]int
		Subjects         map[Action]string
	}
	type args struct {
		action    string
		fieldsMap map[string]string
		extras    map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			"valid",
			fields{
				ts.Templates,
				ts.Actions,
				ts.ReplaceWithField,
				ts.ReplaceWithConst,
				ts.IgnoreFirst,
				ts.Subjects,
			},
			args{},
			"",
			"",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watt := &T{
				Templates:        tt.fields.Templates,
				Actions:          tt.fields.Actions,
				ReplaceWithField: tt.fields.ReplaceWithField,
				ReplaceWithConst: tt.fields.ReplaceWithConst,
				IgnoreFirst:      tt.fields.IgnoreFirst,
				Subjects:         tt.fields.Subjects,
			}
			got, got1, err := watt.Input(tt.args.action, tt.args.fieldsMap, tt.args.extras)
			if (err != nil) != tt.wantErr {
				t.Errorf("T.Input() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("T.Input() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("T.Input() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
