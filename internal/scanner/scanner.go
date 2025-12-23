package scanner

import (
	"bytes"
	"fmt"
	"os"

	"go.uber.org/zap"
)

func Scan(p string, log *zap.Logger) ([]byte, error) {
	const op = "scanner.Scan"

	d, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("%s: read file: %w", op, err)
	}

	if err := processFile(d, log); err != nil {
		return nil, fmt.Errorf("%s: process file: %w", op, err)
	}

	return nil, nil
}

func processFile(d []byte, log *zap.Logger) error {
	const op = "scanner.processFile"
	if err := findConfigs(d, func(cfgData []byte) error {
		curr := cfgData
		for len(curr) > 0 {
			k, v, consumed, err := findKeyValue(curr)
			if err != nil {
				break
			}

			log.Debug("found pair",
				zap.String("string", string(k)),
				zap.String("value", string(v)))
			curr = curr[consumed:]
		}
		return nil
	}); err != nil {
		return fmt.Errorf("%s: find configs: %w", op, err)
	}
	return nil
}

func findConfigs(d []byte, emit func([]byte) error) error {
	const op = "scanner.findConfigs"

	for len(d) > 0 {
		name, conStart, err := findStart(d)
		if err != nil {
			return fmt.Errorf("%s: start idx: %w", op, err)
		}

		conEnd, totalConsumed, err := findEnd(name, d[conStart:])
		if err != nil {
			return fmt.Errorf("%s: end idx: %w", op, err)
		}

		if err := emit(d[conStart : conStart+conEnd]); err != nil {
			return fmt.Errorf("%s: emit func: %w", op, err)
		}

		d = d[conStart+totalConsumed:]
	}

	return nil
}

func findStart(d []byte) (name []byte, nextIdx int, err error) {
	const op = "scanner.findName"

	start := bytes.IndexByte(d, byte('['))
	if start == -1 {
		return nil, 0, fmt.Errorf("%s: start idx: no name start", op)
	}
	end := bytes.IndexByte(d[start:], byte(']'))
	if end == -1 {
		return nil, 0, fmt.Errorf("%s: end idx: no name end", op)
	}

	name = d[start+1 : start+end]

	return name, end + start + 1, nil
}

func findEnd(n []byte, d []byte) (contentEnd int, totalConsumed int, err error) {
	const op = "scanner.findEnd"

	pattern := make([]byte, 2+len(n)+1)
	pattern[0], pattern[1] = '[', '\\'
	copy(pattern[2:], n)
	pattern[2+len(n)] = ']'

	idx := bytes.Index(d, pattern)
	if idx == -1 {
		return 0, 0, fmt.Errorf("%s: start idx: no config end", op)
	}

	return idx, idx + len(pattern), nil
}

func findKeyValue(d []byte) (key []byte, value []byte, contentEnd int, err error) {
	const op = "scanner.findKeyValue"

	start := bytes.Index(d, []byte(":"))
	if start == -1 {
		return nil, nil, 0, fmt.Errorf("%s: start idx: no key value start", op)
	}
	key = d[:start]
	start++

	for start < len(d) && d[start] == ' ' {
		start++
	}

	end := bytes.Index(d[start:], []byte("\n"))
	if end == -1 {
		return nil, nil, 0, fmt.Errorf("%s: end idx: no value end", op)
	}
	value = d[start : end+start]

	return key, value, end + start, nil
}
