package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func isAllUpperAlpha(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	return true
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case 0: // initialized
		consumed, rl, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if consumed == 0 {
			return 0, nil // need more data
		}
		r.RequestLine = rl
		r.state = 1
		return consumed, nil
	case 1: // done
		return 0, fmt.Errorf("parser in done state")
	default:
		return 0, fmt.Errorf("unknown state: %d", r.state)
	}
}

func parseRequestLine(data []byte) (consumed int, rl RequestLine, err error) {
	var i = -1
	i = bytes.Index(data, []byte("\r\n"))

	if i == -1 {
		return 0, RequestLine{}, nil
	}
	line := string(data[:i])
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return 0, RequestLine{}, fmt.Errorf("invalid length of request line")
	}
	if !isAllUpperAlpha(parts[0]) {
		return 0, RequestLine{}, fmt.Errorf("invalid method")
	}
	if !strings.HasPrefix(parts[2], "HTTP/") {
		return 0, RequestLine{}, fmt.Errorf("invalid version prefix")
	}
	vnum := strings.TrimPrefix(parts[2], "HTTP/")
	if vnum != "1.1" {
		return 0, RequestLine{}, fmt.Errorf("unsupported http version: %s", vnum)
	}

	rl = RequestLine{
		HttpVersion:   vnum,
		RequestTarget: parts[1],
		Method:        parts[0],
	}

	return i + 2, rl, nil

}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, 8)
	readTo := 0
	req := &Request{state: 0}

	for req.state != 1 {
		// grow if full
		if readTo == len(buf) {
			nb := make([]byte, len(buf)*2)
			copy(nb, buf[:readTo])
			buf = nb
		}

		// read
		n, err := reader.Read(buf[readTo:])
		if err != nil {
			if err == io.EOF {
				// try parsing whatever we have
				consumed, perr := req.parse(buf[:readTo])
				if perr != nil {
					return nil, perr
				}
				if consumed > 0 {
					copy(buf, buf[consumed:readTo])
					readTo -= consumed
				}
				if req.state != 1 {
					return nil, fmt.Errorf("incomplete request line before EOF")
				}
				break
			}
			return nil, err
		}
		readTo += n

		// parse
		for {
			consumed, perr := req.parse(buf[:readTo])
			if perr != nil {
				return nil, perr
			}
			if consumed == 0 {
				break // need more data
			}
			// shift left
			copy(buf, buf[consumed:readTo])
			readTo -= consumed
			// if done, loop condition will exit
			if req.state == 1 {
				break
			}
		}
	}
	return req, nil
}
