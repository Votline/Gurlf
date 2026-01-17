package core

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"sync"
	"unsafe"

	"github.com/Votline/Gurlf/pkg/scanner"
)

type field struct {
	tag []byte
	idx []int
}
type marshalField struct {
	precomputedTag []byte
	idx            []int
	isConfigName   bool
}
type structCache struct {
	unmFields []field
	marFields []marshalField
	nameIdx   []int
}

var (
	cache    sync.Map
	bufferPool = sync.Pool{
		New: func() any {
			return bytes.NewBuffer(make([]byte, 0, 1024))
		},
	}
)

func Unmarshal(d scanner.Data, v any) error {
	const op = "core.Unmarshal"

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("%s: invalid value: need pointer to value", op)
	}
	rv = rv.Elem()
	rt := rv.Type()

	var info structCache
	if val, ok := cache.Load(rt); ok {
		info = val.(structCache)
	} else {
		info.unmFields = make([]field, 0, rt.NumField())
		fillCache(rt, &info, nil)
		cache.Store(rt, info)
	}

	if len(info.unmFields) == 0 {
		return fmt.Errorf("%s: unmFields unmFields: zero unmFields", op)
	}

	if info.nameIdx != nil {
		if err := setValue(rv.FieldByIndex(info.nameIdx), d.Name); err != nil {
			return fmt.Errorf("%s: set value: %w", op, err)
		}
	}

	for _, ent := range d.Entries {
		key := d.RawData[ent.KeyStart:ent.KeyEnd]
		if len(key) == 0 {
			continue
		}
		val := d.RawData[ent.ValStart:ent.ValEnd]

		for _, f := range info.unmFields {
			if bytes.Equal(f.tag, key) {
				if err := setValue(rv.FieldByIndex(f.idx), val); err != nil {
					return fmt.Errorf("%s: set value: %w", op, err)
				}
			}
		}
	}

	return nil
}

func fillCache(rt reflect.Type, info *structCache, path []int) {
	for i := range rt.NumField() {
		f := rt.Field(i)

		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			path = append(path, i)
			fillCache(f.Type, info, path)
			path = path[:len(path)-1]
			continue
		}

		tag := f.Tag.Get("gurlf")
		if tag == "" {
			continue
		}
		path = append(path, i)

		finalIdx := make([]int, len(path))
		copy(finalIdx, path)

		if tag == "config_name" {
			info.nameIdx = finalIdx
			info.marFields = append(info.marFields, marshalField{
				idx:          finalIdx,
				isConfigName: true,
			})
			path = path[:len(path)-1]
			continue
		}

		info.unmFields = append(info.unmFields, field{
			tag: []byte(tag),
			idx: finalIdx,
		})

		prep := make([]byte, 0, len(tag)+1)
		prep = append(prep, tag...)
		prep = append(prep, ':')
		info.marFields = append(info.marFields, marshalField{
			precomputedTag: prep,
			idx:            finalIdx,
		})

		path = path[:len(path)-1]
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
		v.SetInt(i)
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
	var info structCache
	if val, ok := cache.Load(rt); ok {
		info = val.(structCache)
	} else {
		fillCache(rt, &info, nil)
		cache.Store(rt, info)
	}

	var cfgName []byte
	if info.nameIdx != nil {
		cfgName = appendValue(nil, rv.FieldByIndex(info.nameIdx))
	}

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	res := buf.Bytes()
	var nStart, nEnd int
	for _, f := range info.marFields {
		if rv.Kind() == reflect.Pointer {
			rv = rv.Elem()
		}

		fV := rv.FieldByIndex(f.idx)
		if f.isConfigName {
			continue
		}
		res = append(res, f.precomputedTag...)
		res = appendValue(res, fV)
		res = append(res, '\n')
	}

	if len(cfgName) == 0 {
		final := make([]byte, len(res))
		copy(final, res)
		return final, nil
	}

	finalSize := (len(res) - (nEnd - nStart)) + (len(cfgName) * 2) + 10
	final := make([]byte, 0, finalSize)

	final = append(final, '[')
	final = append(final, cfgName...)
	final = append(final, ']', '\n')

	final = append(final, res...)

	final = append(final, '[', '\\')
	final = append(final, cfgName...)
	final = append(final, ']', '\n', '\n')

	return final, nil
}

func appendValue(dst []byte, v reflect.Value) []byte {
	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if needMultiline(s) {
			dst = append(dst, '`')
			dst = append(dst, s...)
			dst = append(dst, '`')
			return dst
		}
		return append(dst, s...)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.AppendInt(dst, v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.AppendInt(dst, v.Int(), 10)
	case reflect.Float32:
		return strconv.AppendFloat(dst, v.Float(), 'f', -1, 32)
	case reflect.Float64:
		return strconv.AppendFloat(dst, v.Float(), 'f', -1, 64)
	case reflect.Slice:
		b := v.Bytes()
		s := unsafe.String(unsafe.SliceData(b), len(b))
		if needMultiline(s) {
			dst = append(dst, '`')
			dst = append(dst, s...)
			dst = append(dst, '`')
			return dst
		}
		return append(dst, s...)
	}
	return fmt.Append(dst, v.Interface())
}

func needMultiline(s string) bool {
	for i := range len(s) {
		switch s[i] {
			case '\n', '\t', '\r', '`':
				return true
		}
	}
	return false
}

func Encode(wr io.Writer, d []byte) error {
	_, err := wr.Write(d)
	return err
}
