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

	cfgs, err := findConfigs(d)
	if err != nil && len(cfgs) == 0 {
		return nil, fmt.Errorf("%s: find configs: %w", op, err)
	} else if len(cfgs) != 0 {
		log.Warn("find configs error",
			zap.String("op", op),
			zap.Error(err))
	}

	for i, cfg := range cfgs {
		log.Debug("configs!",
			zap.Int("№", i),
			zap.String("cfg", string(cfg)))
	}


	return nil, nil
}

func findConfigs(d []byte) ([][]byte, error) {
	const op = "scanner.findConfigs"

	var cfgs [][]byte

	for len(d) > 0 {
		name, idx, err := findStart(d)
		if err != nil {
			return cfgs, fmt.Errorf("%s: start idx: %w", op, err)
		}

		end, err := findEnd(string(name), d[idx:])
		if err != nil {
			return cfgs, fmt.Errorf("%s: end idx: %w", op, err)
		}

		if end == 0 {
			return nil,
				fmt.Errorf("%s: end idx: zero end idx, invalid config", op)
		}
		cfgs = append(cfgs, d[idx:end])
		d = d[end:]

		fmt.Printf("\nConfig for scan: %s\n", string(d))
	}

	return cfgs, nil
}

func findStart(d []byte) ([]byte, int, error) {
	const op = "scanner.findName"

	start := bytes.IndexByte(d, byte('['))
	if start == -1 {
		return nil, 0, fmt.Errorf("%s: start idx: no name start", op)
	}
	start++

	end := bytes.IndexByte(d[start:], byte(']'))
	if end == -1 {
		return nil, 0, fmt.Errorf("%s: end idx: no name end", op)
	}

	return d[start:end+start], end+start, nil
}

func findEnd(n string, d []byte) (int, error) {
	const op = "scanner.findEnd"

	fmt.Printf("%s", `\`+n)

	end := bytes.Index(d, []byte(`\`+n))
	if end == -1 {
		return 0, fmt.Errorf("%s: end idx: no config end", op)
	}
	return end, nil
}

/*
func Scan(p string, log *zap.Logger) ([]byte, error) {
	const op = "scanner.Scan"

	d, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("%s: open file: %w", op, err)
	}

	name, err := findName(d)
	if err != nil {
		log.Warn("name error",
			zap.String("op", op),
			zap.Error(err))
	}
	log.Debug("find name",
		zap.String("name", string(name)))

	kv, err := findKeyValue(d[len(name):])
	if err != nil {
		log.Warn("key value error",
			zap.String("op", op),
			zap.Error(err))
	}

	for k, v := range kv {
	log.Debug("find key value",
		zap.String("key", k),
		zap.String("val", v))
		
	}

	return nil, nil
}

func findKeyValue(d []byte) (map[string]string, error) {
	const op = "scanner.findKeyValue"

	start := bytes.Index(d, []byte(":"))
	if start == -1 {
		return nil, fmt.Errorf("%s: start idx: no key value start", op)
	}
	key := d[:start]
	start++

	for start < len(d) && d[start] == ' ' {
		start++
	}

	if res := findByQuote(d[start:]); res != nil {
		return res, nil
	}

	end := bytes.Index(d[start:], []byte("\n"))
	if end == -1 { return nil, fmt.Errorf("%s: end idx: no value end", op) }
	value := d[start:end+start]

	res := make(map[string]string, len(key)+len(value))
	res[string(key)] = string(value)
	return res, nil
}

func findByQuote(d []byte) map[string]string {
	start := bytes.Index(d, []byte("`"))
	if start == -1 { return nil }
	return nil
}
*/
