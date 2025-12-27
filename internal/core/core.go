package core

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"

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

	for i := range len(data) {
		cfg := struct {
			ID   int    `gurlf:"ID"`
			Body string `gurlf:"BODY"`
			Headers string `gurlf:"HEADERS"`
		}{}
		if err := c.Unmarshal(data[i], &cfg); err != nil {
			return fmt.Errorf("%s: unmarshal data: %w", op, err)
		}

		fmt.Printf("\nID: %d | Body: %s | Headers: %s\n", cfg.ID, cfg.Body, cfg.Headers)

	}
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
			f := rv.Field(idx)
			val = bytes.TrimSpace(val)

			switch f.Kind() {
			case reflect.String:
				f.SetString(string(val))
			case reflect.Int, reflect.Int64:
				i, err := strconv.ParseInt(string(val), 10, 64)
				if err != nil {
					return fmt.Errorf("%s: cannot parse int: %w", op, err)
				}
				f.SetInt(int64(i))
			case reflect.Slice:
				if f.Type().Elem().Kind() == reflect.Uint8 {
					f.SetBytes(val)
				}
			default:
				return fmt.Errorf("%s: unsupported type: %v", op, f.Kind())
			}
		}
	}

	return nil
}
