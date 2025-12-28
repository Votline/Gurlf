package gurlf

import (
	"github.com/Votline/Gurlf/pkg/core"
	"github.com/Votline/Gurlf/pkg/scanner"
)

func Scan(p string) ([]scanner.Data, error) {
	return scanner.Scan(p)
}

func Unmarshal(d scanner.Data, v any) error {
	return core.Unmarshal(d, v)
}
