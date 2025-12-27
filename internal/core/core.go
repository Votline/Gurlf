package core

import (
	"bytes"
	"fmt"
	"reflect"

	"gurlf/internal/scanner"

	"go.uber.org/zap"
)

type core struct {
	fileName string
	log      *zap.Logger
}

func New(f string, l *zap.Logger) *core {
	return &core{fileName: f, log: l}
}

func (c *core) Start() error {
	const op = "core.Start"

	c.log.Debug("starting process",
		zap.String("file", c.fileName))

	data, err := scanner.Scan(c.fileName, c.log)
	if err != nil {
		return fmt.Errorf("%s: scan file: %w", op, err)
	}

	c.log.Debug("scan complete",
		zap.Int("configs length", len(data)))

	cfg := struct {
		ID   []byte `gurlf:"ID"`
		Body []byte `gurlf:"BODY"`
	}{}
	if err := c.Unmarshal(data[0], &cfg); err != nil {
		return fmt.Errorf("%s: unmarshal data: %w", op, err)
	}

	c.log.Debug("extracted values",
		zap.String("id", string(cfg.ID)),
		zap.String("body", string(cfg.Body)))

	return nil
}

func (c *core) Unmarshal(d scanner.Data, v any) error {
	const op = "core.Unmarshal"

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("%s: invalid value: need pointer to value", op)
	}
	rv = rv.Elem()

	rt := rv.Type()
	cache := make(map[string]int, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		tag := f.Tag.Get("gurlf")
		if tag != "" {
			cache[tag] = i
		}
	}

	for _, ent := range d.Entries {
		key := string(d.RawData[ent.KeyStart:ent.KeyEnd])
		val := d.RawData[ent.ValStart:ent.ValEnd]

		if idx, ok := cache[key]; ok {
			rv.Field(idx).SetBytes(bytes.TrimSpace(val))
		}
	}

	return nil
}
