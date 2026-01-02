package scanner

import (
	"testing"
)

func TestEmit(t *testing.T) {
	cfgData := []byte(`
		ID: 15
		User: dev
		Project: WhereBear
	`)

	tests := []struct {
		key string
		val string
	}{
		{"ID", "15"},
		{"User", "dev"},
		{"Project", "WhereBear"},
	}

	enrs, err := emit(cfgData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(enrs) != len(tests) {
		t.Errorf("expected %d entries, got %d", len(tests), len(enrs))
	}

	for i, ent := range enrs {
		if i >= len(tests) {
			break
		}
		key := cfgData[ent.KeyStart:ent.KeyEnd]
		val := cfgData[ent.ValStart:ent.ValEnd]

		if string(key) != tests[i].key {
			t.Errorf("[%d] key mismatch: expected %q, got %q",
				i, tests[i].key, string(key))
		}
		if string(val) != tests[i].val {
			t.Errorf("[%d] value mismatch: expected %q, got %q",
				i, tests[i].val, string(val))
		}

	}
}

func TestFindConfigs(t *testing.T) {
	cfgData := []byte(`
		[config]
		ID: 15
		[\config]
		[sec]
		ID: 21
		[\sec]
	`)

	res, err := findConfigs(cfgData, func(b []byte) ([]Entry, error) { return nil, nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res) != 2 {
		t.Errorf("len mismatch: expected %d, got %d", 2, len(res))
	}
}

func TestFindStart(t *testing.T) {
	tests := []struct {
		expName string
		expIdx  int
		input   string
	}{
		{
			expName: "config",
			expIdx:  8,
			input:   "[config] hello",
		},
		{
			expName: "new_config",
			expIdx:  12,
			input:   "[new_config] hello",
		},
		{
			expName: "third$ config",
			expIdx:  15,
			input:   "[third$ config] hello",
		},
	}

	for i, tt := range tests {
		actName, actIdx, err := findStart([]byte(tt.input))
		if err != nil && actIdx == -1 {
			t.Fatalf("[%d]: unexpected error: %v", i, err)
		}

		if string(actName) != tt.expName {
			t.Errorf("[%d]: expected %q, got %q",
				i, tt.expName, string(actName))
		}
		if actIdx != tt.expIdx {
			t.Errorf("[%d]: expected %d, got %d",
				i, tt.expIdx, actIdx)
		}
	}
}

func TestFindEnd(t *testing.T) {
	tests := []struct {
		expIdx  int
		expCons int
		input   string
		name    string
	}{
		{
			name:    "config",
			expIdx:  5,
			expCons: 14,
			input:   `some [\config]`,
		},
		{
			name:    "new_config",
			expIdx:  4,
			expCons: 17,
			input:   `sym [\new_config]`,
		},
		{
			name:    "third$ config",
			expIdx:  5,
			expCons: 21,
			input:   `bols [\third$ config]`,
		},
	}

	for i, tt := range tests {
		actIdx, actCons, err := findEnd([]byte(tt.name), []byte(tt.input))
		if err != nil && i <= len(tests) {
			t.Fatalf("[%d]: unexpected error: %v", i, err)
		}

		if actIdx != tt.expIdx {
			t.Errorf("[%d]: expected idx %d, got %d",
				i, tt.expIdx, actIdx)
		}
		if actCons != tt.expCons {
			t.Errorf("[%d]: expected cons %d, got %d",
				i, tt.expCons, actCons)
		}
	}
}

func TestFindKeyValue(t *testing.T) {
	tests := []struct {
		input  string
		expKey string
		expVal string
	}{
		{"ID: 115\n", "ID", "115"},
		{"Hdrs: Content-type\n", "Hdrs", "Content-type"},
		{"Body: `115 road\n`", "Body", "115 road\n"},
		{"Cks: `Maref`", "Cks", "Maref"},
	}

	for i, tt := range tests {
		kS, kE, vS, vE, _, err := findKeyValue([]byte(tt.input))

		if err != nil {
			t.Fatalf("[%d]: unexpected error: %v", i, err)
		}

		actKey := string(tt.input[kS:kE])
		actVal := string(tt.input[vS:vE])

		if actKey != tt.expKey {
			t.Errorf("[%d]: expected idx %q, got %q",
				i, tt.expKey, actKey)
		}
		if actVal != tt.expVal {
			t.Errorf("[%d]: expected cons %q, got %q",
				i, tt.expVal, actVal)
		}
	}
}
