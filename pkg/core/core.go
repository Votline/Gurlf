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
	cache := make(map[string][]int, rt.NumField())
	fillCache(rt, cache, nil)

	if len(cache) == 0 {
		return fmt.Errorf("%s: cache fields: zero fields", op)
	}

	if idx, ok := cache["config_name"]; ok {
		f := rv.FieldByIndex(idx)

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
			f := rv.FieldByIndex(idx)
			val = bytes.TrimSpace(val)

			if err := setValue(f, val); err != nil {
				return fmt.Errorf("%s: set value: %w", op, err)
			}
		}
	}

	return nil
}

func fillCache(rt reflect.Type, cache map[string][]int, baseIdx []int) {
	for i := range rt.NumField() {
		f := rt.Field(i)
		curIdx := append(append([]int{}, baseIdx...), i)

		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			fillCache(f.Type, cache, curIdx)
			continue
		}

		tag := f.Tag.Get("gurlf")
		if tag == "" {
			tag = f.Name
		}
		cache[tag] = curIdx
	}
}

func setValue(v reflect.Value, val []byte) error {
	const op = "core.setValue"

	switch v.Kind() {
	case reflect.String:
		str := unsafe.String(unsafe.SliceData(val), len(val))
		v.SetString(str)
	case reflect.Int, reflect.Int32, reflect.Int64:
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

	res := make([]byte, 0, 1024)
	fields := make([]byte, 0, 512)
	var cfgName []byte

	writeRecursive(rv, &fields, &cfgName)

	if len(cfgName) != 0 {
		res = append(res, '[')
		res = append(res, cfgName...)
		res = append(res, ']', '\n')
	}

	res = append(res, fields...)

	if len(cfgName) != 0 {
		res = append(res, '[', '\\')
		res = append(res, cfgName...)
		res = append(res, ']', '\n')
	}

	return res, nil
}

func writeRecursive(rv reflect.Value, dst *[]byte, name *[]byte) {
	rt := rv.Type()
	for i := range rv.NumField() {
		fV, fT := rv.Field(i), rt.Field(i)

		if !fV.CanInterface() { continue }

		if fT.Anonymous && fV.Kind() == reflect.Struct {
			writeRecursive(fV, dst, name)
			continue
		}

		tag := fT.Tag.Get("gurlf")
		if tag == "" {
			tag = fT.Name
		}

		if tag == "config_name" {
			tmp := []byte("")
			*name = appendValue(&tmp, fV)
			continue
		}

		*dst = append(*dst, tag...)
		*dst = append(*dst, ':')
		*dst = appendValue(dst, fV)
		*dst = append(*dst, '\n')
	}
}
func appendValue(dst *[]byte, v reflect.Value) []byte {
	switch v.Kind() {
	case reflect.String:
		return append(*dst, v.String()...)
	case reflect.Int, reflect.Int32, reflect.Int64:
		return strconv.AppendInt(*dst, v.Int(), 10)
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return append(*dst, v.Bytes()...)
		}
	}
	return fmt.Append(*dst, v.Interface())
}

func Encode(wr io.Writer, d []byte) error {
	_, err := wr.Write(d)
	return err
}
