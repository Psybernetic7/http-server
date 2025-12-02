package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/Psybernetic7/http-server/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	requestStateParsingRequestLine = iota
	requestStateParsingHeaders
	requestStateDone
)

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
	total := 0
	for {
		if len(data[total:]) == 0 || r.state == requestStateDone {
			return total, nil
		}
		n, err := r.parseSingle(data[total:])
		if err != nil {
			return total, err
		}
		if n == 0 {
			return total, nil
		}
		total += n
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
	req := &Request{
		state:   requestStateParsingRequestLine,
		Headers: headers.NewHeaders(),
	}

	for req.state != requestStateDone {
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
				if req.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request before EOF")
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
			if req.state == requestStateDone {
				break
			}
		}
	}
	return req, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateParsingRequestLine:
		consumed, rl, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if consumed == 0 {
			return 0, nil
		}
		r.RequestLine = rl
		r.state = requestStateParsingHeaders
		return consumed, nil

	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		if done {
			r.state = requestStateDone
		}
		return n, nil

	default:
		return 0, fmt.Errorf("unknown state: %d", r.state)
	}
}
