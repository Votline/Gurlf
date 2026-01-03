package core

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Votline/Gurlf/pkg/scanner"
)

func TestUnmarshal(t *testing.T) {
	raw := []byte(`
		[config]
		ID:12
		User:admin
		Encoder:console
		[\config]`)

	testData := scanner.Data{
		Name:    []byte("config"),
		RawData: raw,
		Entries: []scanner.Entry{
			{
				KeyStart: 14, KeyEnd: 16,
				ValStart: 17, ValEnd: 19,
			},
			{
				KeyStart: 22, KeyEnd: 26,
				ValStart: 27, ValEnd: 32,
			},
			{
				KeyStart: 35, KeyEnd: 42,
				ValStart: 43, ValEnd: 50,
			},
		},
	}

	res := struct {
		Name string `gurlf:"config_name"`
		ID   int    `gurlf:"ID"`
		User string `gurlf:"User"`
		Enc  string `gurlf:"Encoder"`
	}{}

	if err := Unmarshal(testData, &res); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		expected string
		actual   string
	}{
		{"12", fmt.Sprint(res.ID)},
		{"admin", res.User},
		{"console", res.Enc},
		{"config", res.Name},
	}

	for i, tt := range tests {
		if tt.actual != tt.expected {
			t.Errorf("[%d]: expected %q, got %q",
				i, tt.expected, tt.actual)
		}
	}
}

func TestMarshal(t *testing.T) {
	type testData struct {
		ID   int    `gurlf:"ID"`
		User string `gurlf:"User"`
	}

	tests := []struct {
		input testData
		want  map[string]string
	}{
		{
			input: testData{12, "admin"},
			want:  map[string]string{"ID": "12", "User": "admin"},
		},
		{
			input: testData{45, "dev"},
			want:  map[string]string{"ID": "45", "User": "dev"},
		},
	}

	for i, tt := range tests {
		b, err := Marshal(tt.input)
		if err != nil {
			t.Fatalf("[%d]: unexpected error: %v", i, err)
		}

		got := make(map[string]string)
		lines := strings.Split(strings.TrimSpace(string(b)), "\n")

		for j, line := range lines {
			kv := strings.SplitN(line, ":", 2)
			if len(kv) != 2 {
				t.Errorf("[%d]: in lines [%d]: Invalid line format: %q",
					i, j, line)
				continue
			}
			got[kv[0]] = kv[1]
		}

		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("[%d]:\n Got: %v\nWant: %v", i, got, tt.want)
		}
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"ID: 15\nBody:15\t21\n\n"},
	}

	for i, tt := range tests {
		buf := new(bytes.Buffer)
		if err := Encode(buf, []byte(tt.input)); err != nil {
			t.Fatalf("[%d]: unexpected error: %v", i, err)
		}
		if buf.String() != tt.input {
			t.Errorf("[%d]: expected %q, got %q",
				i, buf.String(), tt.input)
		}
		if buf.Len() != len(tt.input) {
			t.Errorf("[%d]: expected len %d, got %d",
				i, buf.Len(), len(tt.input))

		}
	}
}

func TestInlineUnmarshal(t *testing.T) {
	type Base struct {
		Enable int `gurlf:"enable"`
	}
	type Config struct {
		Base
		ID string `gurlf:"id"`
		Name string `gurlf:"config_name"`
	}
	data := scanner.Data{
		Name: []byte("current"),
		RawData: []byte("enable:0\nid:115"),
		Entries: []scanner.Entry{
			{KeyStart: 0, KeyEnd: 6, ValStart: 7, ValEnd: 8},
			{KeyStart: 9, KeyEnd: 11, ValStart: 12, ValEnd: 15},
		},
	}
	var cfg Config
	if err := Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		expected string
		actual   string
	}{
		{"115", fmt.Sprint(cfg.ID)},
		{"0", fmt.Sprint(cfg.Enable)},
		{"current", cfg.Name},
	}

	for i, tt := range tests {
		if tt.actual != tt.expected {
			t.Errorf("[%d]: expected %q, got %q",
				i, tt.expected, tt.actual)
		}
	}
}

func TestInlineMarshal(t *testing.T) {
	type Base struct {
		Version string `gurlf:"v"`
	}
	type Config struct{
		Base
		ID string `gurlf:"id"`
		Name string `gurlf:"config_name"`
	}
	c := Config{
		Base: Base{Version: "15"},
		ID: "current",
		Name: "cfg",
	}
	got, err := Marshal(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	res := string(got)
	expectedLines := []string{"cfg", "15", "current"}
	for i, line := range expectedLines {
		if !bytes.Contains(got, []byte(line)) {
			t.Errorf("[%d]: expected %q, got %q",
				i, line, res)
		}
	}
}
