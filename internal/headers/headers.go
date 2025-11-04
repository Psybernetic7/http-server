package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 2, true, nil
	}
	line := data[:idx]
	line = bytes.TrimSpace(line)
	colIdx := bytes.Index(line, []byte(":"))
	if colIdx == -1 {
		return 0, false, fmt.Errorf("missing colon")
	}
	if colIdx > 0 && line[colIdx-1] == ' ' {
		return 0, false, fmt.Errorf("space before colon")
	}
	key := line[:colIdx]
	value := line[colIdx+1:]
	key = bytes.TrimSpace(key)
	value = bytes.TrimSpace(value)
	if len(key) == 0 {
		return 0, false, fmt.Errorf("empty key")
	}
	for _, b := range key {
		if !isValidHeaderChar(byte(b)) {
			return 0, false, fmt.Errorf("invalid header key character: %q", b)
		}
	}
	key = bytes.ToLower(key)
	h[string(key)] = string(value)
	return idx + 2, false, nil

}

func isValidHeaderChar(b byte) bool {
	// letters and digits
	if ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') || ('0' <= b && b <= '9') {
		return true
	}
	// allowed specials: ! # $ % & ' * + - . ^ _ ` | ~
	switch b {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}
