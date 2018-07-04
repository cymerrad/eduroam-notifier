package ts

import (
	"eduroam-notifier/app/models"
	"reflect"
	"testing"
	"text/template"
)

const LOGIN_INCORRECT = "Login incorrect (mschap: MS-CHAP2-Response is incorrect)"

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
			map[Action]TemplateID{LOGIN_INCORRECT: "wrong_password"},
			map[TemplateTag]Field{"mac": "source-mac", "pesel": "Pesel"},
			map[TemplateTag]ConstValue{"signature": "DSK UW"},
			map[Action]int{LOGIN_INCORRECT: 5},
			map[Action]string{LOGIN_INCORRECT: "Ostrzeżenie Eduroam"},
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
				Value: GenerateJSON(OnAction, LOGIN_INCORRECT, DoActionIgnoreFirstN, "abc"),
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
			args{
				LOGIN_INCORRECT,
				map[string]string{
					"EDUROAM_ACT":      ": (1548821)  Login incorrect (mschap: MS-CHAP2-Response is incorrect): [71072700875@uw.edu.pl] (from client trapeze-mx1 port 56454 cli 9C-F3-87-1E-DD-37",
					"Pesel":            "71072700875",
					"Realm":            "uw.edu.pl",
					"USERNAME":         "trapeze-mx1",
					"Username":         "71                                                                                                010.030.061.024.41245-010.012.003.236.00080: 072700875",
					"WINDOWSMAC":       "9C-F3-87-1E-DD-37",
					"action":           LOGIN_INCORRECT,
					"client":           "trapeze-mx1",
					"facility":         "local1",
					"gl2_remote_ip":    "10.30.87.42",
					"gl2_remote_port":  "0",
					"gl2_source_input": "55512e88e4b02f16ad5339c7",
					"gl2_source_node":  "64f19870-4111-42dd-aef2-e7d662535efb",
					"level":            "0",
					"source-mac":       "9C-F3-87-1E-DD-37",
					"source-user":      "71072700875@uw.edu.pl",
				},
				map[string]string{
					"CANCEL_LINK":    "\u003ca href=\"http://localhost:9000/cancel/11fc9b482b5120a3c5a840e193039105df55ae7588b5b6854a278b8cf112586c\"\u003eClick me\u003c/a\u003e",
					"COUNT_MAC":      "3",
					"COUNT_PESEL":    "3",
					"COUNT_USERNAME": "3",
					"FIRST_NAME":     "Jarosław",
					"NATIONALITY":    "PL",
					"SECOND_NAME":    "Leonard",
					"SEX":            "M",
					"SURNAME":        "Jakielaszek",
				},
			},
			`Witam.
Użytkowniku o numerze pesel 71072700875 próbowałeś zalogować się z urządzenia 9C-F3-87-1E-DD-37, ale wprowadziłeś złe hasło po raz 3.
Jeżeli nie chcesz otrzymywać więcej takich maili, kliknij w <a href="http://localhost:9000/cancel/11fc9b482b5120a3c5a840e193039105df55ae7588b5b6854a278b8cf112586c">Click me</a>.

Z poważaniem,
DSK UW`,
			`Ostrzeżenie Eduroam`,
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
