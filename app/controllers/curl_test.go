package controllers

import (
	"net/url"
	"reflect"
	"testing"
)

func Test_getAllTemplateKeys(t *testing.T) {
	type args struct {
		form url.Values
	}
	tests := []struct {
		name string
		args args
		want map[string]int
	}{
		{"hello there", args{url.Values{"template1": []string{"twoja mama"}}}, map[string]int{"template1": 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAllTemplateKeys(tt.args.form); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAllTemplateKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
