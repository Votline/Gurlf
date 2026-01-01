package core

import (
	"reflect"
	"fmt"
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
		Name string `gurlf:"name"`
		ID   int    `gurlf:"ID"`
		User string `gurlf:"User"`
		Enc  string `gurlf:"Encoder"`
	}{}

	if err := Unmarshal(testData, &res); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct{
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
