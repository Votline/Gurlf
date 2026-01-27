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

type Scanner struct {
	enBuf []Entry
	dtBuf []Data
}

var ScannerPool = sync.Pool{
	New: func() any {
		return &Scanner{
			enBuf: make([]Entry, 0, 256),
			dtBuf: make([]Data, 0, 32),
		}
	},
}

func (s *Scanner) Scan(d []byte) ([]Data, error) {
	const op = "scanner.Scan"

	s.enBuf = s.enBuf[:0]
	s.dtBuf = s.dtBuf[:0]
	for len(d) > 0 {
		name, conStart, err := findStart(d)
		if err != nil {
			return nil, fmt.Errorf("%s: start idx: %w", op, err)
		} else if name == nil {
			return s.dtBuf, nil
		}

		conEnd, totalConsumed, err := findEnd(name, d[conStart:])
		if err != nil {
			return nil, fmt.Errorf("%s: end idx: %w", op, err)
		}

		start := len(s.enBuf)
		s.emit(d[conStart : conStart+conEnd])
		end := len(s.enBuf)

		s.dtBuf = append(s.dtBuf, Data{
			Name:    name,
			RawData: d[conStart : conStart+conEnd],
			Entries: s.enBuf[start:end],
		})

		d = d[conStart+totalConsumed:]
	}

	res := make([]Data, len(s.dtBuf))
	copy(res, s.dtBuf)
	s.dtBuf = s.dtBuf[:0]

	return res, nil
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

	for i := 0; i+2+len(n) <= len(d); i++ {
		if d[i] == '[' && d[i+1] == '\\' {
			if bytes.Equal(d[i+2:i+2+len(n)], n) && i+2+len(n) < len(d) && d[i+2+len(n)] == ']' {
				return i, i + 3 + len(n), nil
			}
		}
	}

	return -1, -1, fmt.Errorf("%s: no end", op)
}

func (s *Scanner) emit(cfgData []byte) {
	offset := 0
	curr := cfgData
	for len(curr) > 0 {
		kS, kE, vS, vE, consumed, err := findKeyValue(curr)
		if err != nil {
			break
		}

		s.enBuf = append(s.enBuf, Entry{
			KeyStart: kS + offset, KeyEnd: kE + offset,
			ValStart: vS + offset, ValEnd: vE + offset,
		})

		curr = curr[consumed:]
		offset += consumed
	}
}

func findKeyValue(d []byte) (keyS, keyE, valS, valE int, contentEnd int, err error) {
	const op = "scanner.findKeyValue"

	start := bytes.IndexByte(d, ':')
	if start == -1 {
		return 0, 0, 0, 0, 0, fmt.Errorf("%s: start idx: no key value start", op)
	}
	seg := d[:start]

	i := 0
	for i < len(seg) && isSpace(seg[i]) {
		i++
	}
	keyS = i

	j := len(seg) - 1
	for j < len(seg) && isSpace(seg[j]) {
		j--
	}
	keyE = j + 1

	start++
	for start < len(d) && d[start] == ' ' {
		start++
	}

	if start+1 < len(d) && d[start] == '`' {
		valE, firstAny := -1, -1
		for i := start + 1; i < len(d); i++ {
			if d[i] == '`' {
				if d[i-1] == '\n' && (i+1 >= len(d) || d[i+1] == '\n' || d[i+1] == '\r') {
					valE = i
					break
				}
				if firstAny == -1 {
					firstAny = i
				}
			}
		}
		if valE == -1 {
			valE = firstAny
		}

		valS = start + 1
		lineEnd := bytes.IndexByte(d[valE:], '\n')

		if lineEnd == -1 {
			return keyS, keyE, valS, valE, valE, nil
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
