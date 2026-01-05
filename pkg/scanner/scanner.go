package scanner

import (
	"bytes"
	"fmt"
	"sync"
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

var entryPool = sync.Pool{
	New: func() any {
		b := make([]Entry, 0, 512)
		return &b
	},
}

var dataPool = sync.Pool{
	New: func() any {
		b := make([]Data, 0, 32)
		return &b
	},
}

func Scan(d []byte) ([]Data, error) {
	const op = "scanner.Scan"

	enPtr := entryPool.Get().(*[]Entry)
	dtPtr := dataPool.Get().(*[]Data)
	enBuf := (*enPtr)[:0]
	dtBuf := (*dtPtr)[:0]
	
	cfgs, err := findConfigs(d, enBuf, dtBuf)

	*enPtr = enBuf
	*dtPtr = dtBuf
	defer entryPool.Put(enPtr)
	defer dataPool.Put(dtPtr)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	res := make([]Data, len(cfgs))
	copy(res, cfgs)

	return res, nil
}

func emit(cfgData []byte, buf []Entry) ([]Entry, error) {
	offset := 0
	curr := cfgData
	enrs := buf[:0]
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
	}
	return enrs, nil
}

func findConfigs(d []byte, entBuf []Entry, dataBuf []Data) ([]Data, error) {
	const op = "scanner.findConfigs"

	entOffset := 0
	dataBuf = (dataBuf)[:0]
	fullEntBuf := (entBuf)[:cap(entBuf)]
	for len(d) > 0 {
		name, conStart, err := findStart(d)
		if err != nil {
			return nil, fmt.Errorf("%s: start idx: %w", op, err)
		} else if name == nil {
			return dataBuf, nil
		}

		conEnd, totalConsumed, err := findEnd(name, d[conStart:])
		if err != nil {
			return nil, fmt.Errorf("%s: end idx: %w", op, err)
		}

		enrs, err := emit(d[conStart : conStart+conEnd], fullEntBuf[entOffset:])
		if err != nil {
			return nil, fmt.Errorf("%s: emit func: %w", op, err)
		}

		cfg := Data {
			Name: name,
			RawData: d[conStart : conStart+conEnd],
			Entries: enrs,
		}

		dataBuf = append(dataBuf, cfg)
		entOffset += len(enrs)

		d = d[conStart+totalConsumed:]
	}

	return dataBuf, nil
}

func findStart(d []byte) (name []byte, nextIdx int, err error) {
	const op = "scanner.findName"

	i := 0
	for i < len(d) && isSpace(d[i]) {
	    i++
	}
	if i == len(d) {
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

	start := bytes.IndexByte(d, ':')
	if start == -1 {
		return 0, 0, 0, 0, 0, fmt.Errorf("%s: start idx: no key value start", op)
	}
	seg := d[:start]

	i:=0
	for i < len(seg) && isSpace(seg[i]) {
		i++
	}
	keyS = i

	j:=len(seg)-1
	for i < len(seg) && isSpace(seg[i]) {
		j--
	}
	keyE = j+1

	start++
	for start < len(d) && d[start] == ' ' {
		start++
	}

	if start+1 < len(d) && d[start] == '`' {
		vE := bytes.IndexByte(d[start+1:], '`')
		if valE == -1 {
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

	end := bytes.Index(d[start:], []byte("\n"))
	if end == -1 {
		return 0, 0, 0, 0, 0,
			fmt.Errorf("%s: end idx: no value end", op)
	}

	valS = start
	valE = end + start

	return keyS, keyE, valS, valE, end + start + 1, nil
}

func isSpace(r byte) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\v' || r == '\f'
}
