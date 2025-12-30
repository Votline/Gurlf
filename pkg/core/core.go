package core

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"

	"github.com/Votline/Gurlf/pkg/scanner"
)

func Unmarshal(d scanner.Data, v any) error {
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

	if len(cache) <= 0 {
		return fmt.Errorf("%s: cache fields: zero fields", op)
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
