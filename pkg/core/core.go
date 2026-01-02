package core

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"unsafe"

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

	if idx, ok := cache["config_name"]; ok {
		f := rv.Field(idx)

		if err := setValue(f, d.Name); err != nil {
			return fmt.Errorf("%s: set value: %w", op, err)
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

			if err := setValue(f, val); err != nil {
				return fmt.Errorf("%s: set value: %w", op, err)
			}
		}
	}

	return nil
}

func setValue(v reflect.Value, val []byte) error {
	const op = "core.setValue"

	switch v.Kind() {
	case reflect.String:
		str := unsafe.String(unsafe.SliceData(val), len(val))
		v.SetString(str)
	case reflect.Int, reflect.Int64:
		str := unsafe.String(unsafe.SliceData(val), len(val))
		i, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fmt.Errorf("%s: cannot parse int: %w", op, err)
		}
		v.SetInt(int64(i))
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			v.SetBytes(val)
		}
	default:
		return fmt.Errorf("%s: unsupported type: %v", op, v.Kind())
	}

	return nil
}

func Marshal(v any) ([]byte, error) {
	const op = "core.Marshal"

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s: invalid value: need struct, but got %q",
			op, rv.Kind())
	}

	rt := rv.Type()
	m := make(map[string]any, rv.NumField())
	for i := range rv.NumField() {
		fieldT := rt.Field(i)
		fieldV := rv.Field(i)

		if !fieldV.CanInterface() {
			continue
		}

		key := fieldT.Tag.Get("gurlf")
		if key == "" {
			key = fieldT.Name
		}

		m[key] = fieldV.Interface()
	}

	name := m["config_name"]
	delete(m, "config_name")
	var b bytes.Buffer
	if err := writeName(nil, name, &b); err != nil {
		return nil, err
	}
	for k, v := range m {
		b.WriteString(k)
		b.WriteByte(':')
		b.WriteString(fmt.Sprint(v))
		b.WriteByte('\n')
	}
	if err := writeName([]byte(`\`), name, &b); err != nil {
		return nil, err
	}
	res := b.Bytes()

	return res, nil
}

func writeName(prefix []byte, n any, b *bytes.Buffer) error {
	const op = "core.writeName"

	switch val := n.(type){
	case string:
		if val == "" { return nil }
		b.Write(prefix)
		b.WriteString(val)
	case []byte:
		if len(val) == 0 { return nil }
		b.Write(prefix)
		b.Write(val)
	case int:
		b.Write(prefix)
		b.WriteString(strconv.Itoa(val))
	case rune:
		b.Write(prefix)
		b.WriteRune(val)
	case nil:
		return nil
	default:
		return fmt.Errorf("%s: unsupported type: %v", op, val)
	}

	return nil
}

func Encode(wr io.Writer, d []byte) error {
	_, err := wr.Write(d)
	return err
}
