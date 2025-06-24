package model

import (
	"testing"
)

func TestParseForm(t *testing.T) {
	tests := []struct {
		input  string
		want   string
		wantOk bool
	}{
		{input: ""},       // empty
		{input: "tablte"}, // misspell
		{input: "tablet", want: "tablet", wantOk: true},    // raw
		{input: " tablEt\t", want: "tablet", wantOk: true}, // whitespace and capital
		{input: "Capsule ", want: "capsule", wantOk: true},
		{input: " liquiD", want: "liquid", wantOk: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, gotOk := ParseForm(tt.input)
			if string(got) != tt.want {
				t.Errorf("ParseForm() got = %v, want %v", string(got), tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("ParseForm() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
