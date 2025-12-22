package scanner

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

func Scan(p string, log *zap.Logger) ([]byte, error) {
	const op = "scanner.Scan"

	d, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("%s: read file: %w", op, err)
	}

	cfgs, err := findConfigs(d)
	if err != nil && len(cfgs) == 0 {
		return nil, fmt.Errorf("%s: find configs: %w", op, err)
	} else if len(cfgs) != 0 {
		log.Warn("find configs error",
			zap.String("op", op),
			zap.Error(err))
	}

	for i, cfg := range cfgs {
		var (
			err error
			k []byte
			v []byte
			conEnd int
		)
		for err == nil {
			k, v, conEnd, err = findKeyValue(cfg)
			log.Debug("configs!",
				zap.Int("№", i),
				zap.String("key", string(k)),
				zap.String("value", string(v)))
			cfg = cfg[conEnd:]
			time.Sleep(10*time.Millisecond)
		}

		log.Warn("find key value err",
			zap.String("op", op),
			zap.Error(err))
	}

	return nil, nil
}

func findConfigs(d []byte) ([][]byte, error) {
	const op = "scanner.findConfigs"

	var cfgs [][]byte

	for len(d) > 0 {
		name, conStart, err := findStart(d)
		if err != nil {
			return cfgs, fmt.Errorf("%s: start idx: %w", op, err)
		}
		conEnd, totalConsumed, err := findEnd(string(name), d[conStart:])
		if err != nil {
			return cfgs, fmt.Errorf("%s: end idx: %w", op, err)
		}
		
		cfgs = append(cfgs, d[conStart:conStart+conEnd])
		d = d[conStart+totalConsumed:]
	}

	return cfgs, nil
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

func findEnd(n string, d []byte) (contentEnd int, totalConsumed int, err error) {
	const op = "scanner.findEnd"

	pattern := []byte(`[\` + n + `]`)
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
	value = d[start:end+start]

	return key, value, end+start, nil
}
