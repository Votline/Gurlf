package scanner

import (
	"bytes"
	"fmt"
	"os"

	"go.uber.org/zap"
)

type Entry struct {
	KeyStart, KeyEnd int
	ValStart, ValEnd int
}

type Data struct {
	Name    []byte
	RawData []byte
	Entries []Entry
}

func Scan(p string, log *zap.Logger) ([]Data, error) {
	const op = "scanner.Scan"

	d, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("%s: read file: %w", op, err)
	}

	cfgs, err := processFile(d, log)
	if err != nil {
		return nil, fmt.Errorf("%s: process file: %w", op, err)
	}

	return cfgs, nil
}

func processFile(d []byte, log *zap.Logger) ([]Data, error) {
	const op = "scanner.processFile"
	cfgs, err := findConfigs(d, func(cfgData []byte) ([]Entry, error) {
		offset := 0
		curr := cfgData
		var enrs []Entry
		for len(curr) > 0 {
			kS, kE, vS, vE, consumed, err := findKeyValue(curr)
			if err != nil {
				break
			}

			enrs = append(enrs, Entry{
				KeyStart: kS + offset, KeyEnd: kE + offset,
				ValStart: vS + offset, ValEnd: vE + offset,
			})

			curr = curr[consumed:]
			offset += consumed

			log.Debug("extracted indexes",
				zap.String("op", op),
				zap.Int("key start", kS),
				zap.Int("key end", kE),
				zap.Int("value start", vS),
				zap.Int("value end", vE),
				zap.Int("consumed", consumed))
		}
		return enrs, nil
	}, log)
	if err != nil {
		return nil, fmt.Errorf("%s: find configs: %w", op, err)
	}
	return cfgs, nil
}

func findConfigs(d []byte, emit func([]byte) ([]Entry, error), log *zap.Logger) ([]Data, error) {
	const op = "scanner.findConfigs"

	var cfgs []Data

	for len(d) > 0 {
		name, conStart, err := findStart(d)
		if err != nil {
			return cfgs, fmt.Errorf("%s: start idx: %w", op, err)
		} else if name == nil {
			return cfgs, nil
		}

		conEnd, totalConsumed, err := findEnd(name, d[conStart:])
		if err != nil {
			return cfgs, fmt.Errorf("%s: end idx: %w", op, err)
		}

		enrs, err := emit(d[conStart : conStart+conEnd])
		if err != nil {
			return cfgs, fmt.Errorf("%s: emit func: %w", op, err)
		}

		var cfg Data
		cfg.Name = name
		cfg.RawData = d[conStart : conStart+conEnd]
		cfg.Entries = enrs
		cfgs = append(cfgs, cfg)

		log.Debug("new config",
			zap.Int("len data", len(cfg.RawData)),
			zap.Int("len entries", len(cfg.Entries)))

		d = d[conStart+totalConsumed:]
	}

	return cfgs, nil
}

func findStart(d []byte) (name []byte, nextIdx int, err error) {
	const op = "scanner.findName"

	if bytes.TrimSpace(d) == nil {
		return nil, 0, nil
	}

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

func findKeyValue(d []byte) (keyS, keyE, valS, valE int, contentEnd int, err error) {
	const op = "scanner.findKeyValue"

	start := bytes.Index(d, []byte(":"))
	if start == -1 {
		return 0, 0, 0, 0, 0, fmt.Errorf("%s: start idx: no key value start", op)
	}
	seg := d[:start]
	keyS = bytes.IndexFunc(seg, func(r rune) bool {
		return !isSpace(r)
	})
	last := bytes.LastIndexFunc(seg, func(r rune) bool {
		return !isSpace(r)
	})
	keyE = last + 1
	start++

	for start < len(d) && d[start] == ' ' {
		start++
	}

	end := bytes.Index(d[start:], []byte("\n"))
	if end == -1 {
		return 0, 0, 0, 0, 0,
			fmt.Errorf("%s: end idx: no value end", op)
	}

	if start+1 < len(d) && d[start] == '`' {
		vE, err := findByQuote(d[start+1:])
		if err != nil {
			return 0, 0, 0, 0, 0,
				fmt.Errorf("%s: quote end idx: no value end", op)
		}
		valS, valE = start+1, vE+start+1
		lineEnd := bytes.IndexByte(d[valE-1:], '\n')

		if lineEnd == -1 {
			return keyS, keyE, valS, valE, vE + start, nil
		}

		return keyS, keyE, valS, valE, valE + lineEnd, nil
	}

	valS = start
	valE = end + start

	return keyS, keyE, valS, valE, end + start + 1, nil
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\v' || r == '\f'
}

func findByQuote(d []byte) (valE int, err error) {
	const op = "scanner.findByQuote"
	valE = bytes.IndexByte(d, '`')
	if valE == -1 {
		return -1, fmt.Errorf("%s: end idx: no value end", op)
	}

	return valE, nil
}
