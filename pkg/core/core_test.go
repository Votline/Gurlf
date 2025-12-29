package core

import (
	"github.com/Votline/Gurlf/pkg/scanner"
	"testing"
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
		ID   int    `gurlf:"ID"`
		User string `gurlf:"User"`
		Enc  string `gurlf:"Encoder"`
	}{}

	if err := Unmarshal(testData, &res); err != nil {
		t.Fatal("Unmarshal error: %w", err)
	}

	if res.ID != 12 {
		t.Errorf("Expected: '12', received: %d", res.ID)
	}
	if res.User != "admin" {
		t.Errorf("Expected: 'admin', received: %s", res.User)
	}
	if res.Enc != "console" {
		t.Errorf("Expected: 'console', received: %s", res.Enc)
	}

}
