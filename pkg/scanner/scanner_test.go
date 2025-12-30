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

	type expected struct {
		key string
		val string
	}
	tests := []expected{
		{"ID", "15"},
		{"User", "dev"},
		{"Project", "WhereBear"},
	}

	enrs, err := emit(cfgData)
	if err != nil {
		t.Fatal("Emit error: ", err)
	}

	if len(enrs) != len(tests) {
		t.Errorf("expected %d entries, got %d", len(tests), len(enrs))
	}

	for i, ent := range enrs {
		if i >= len(tests) { break }
		key := cfgData[ent.KeyStart : ent.KeyEnd]
		val := cfgData[ent.ValStart : ent.ValEnd]

		if string(key) != tests[i].key {
			t.Errorf("[%d] key mismatch: expected %s, got %s",
				i, tests[i].key, string(key))
		}
		if string(val) != tests[i].val {
			t.Errorf("[%d] value mismatch: expected %s, got %s",
				i, tests[i].val, string(val))
		}

	}
}
