package gurlf

import (
	"io"
	"os"

	"github.com/Votline/Gurlf/pkg/core"
	"github.com/Votline/Gurlf/pkg/scanner"
)

func Scan(d []byte) ([]scanner.Data, error) {
	return scanner.Scan(d)
}

func ScanFile(p string) ([]scanner.Data, error) {
	d, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	return scanner.Scan(d)
}

func Unmarshal(d scanner.Data, v any) error {
	return core.Unmarshal(d, v)
}

func Encode(wr io.Writer, d []byte) error {
	return core.Encode(wr, d)
}

func EncodeFile(p string, d []byte) error {
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	return core.Encode(f, d)
}
