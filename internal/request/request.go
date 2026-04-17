package request

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
)

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	req := &Request{state: requestStateInitialized}
	readToIndex := 0

	for req.state != requestStateDone {
		numBytesRead, err := reader.Read(buf[readToIndex:])
		readToIndex += numBytesRead
		if readToIndex == cap(buf) {
			newBuf := make([]byte, cap(buf)*2)
			copy(newBuf, buf[:readToIndex])
			buf = newBuf
		}

		numBytesParsed, parseErr := req.parse(buf[:readToIndex])
		if parseErr != nil {
			return nil, parseErr
		}

		// Shifts the unconsumed bytes to the front of the buffer
		copy(buf, buf[numBytesParsed:readToIndex])
		readToIndex -= numBytesParsed

		if err == io.EOF {
			if req.state != requestStateDone {
				return nil, fmt.Errorf("incomplete request: unexpected EOF in state %d", req.state)
			}
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		rl, consumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if consumed == 0 {
			return 0, nil
		}

		r.RequestLine = *rl
		r.Headers = make(map[string]string)
		r.state = requestStateParsingHeaders

		return consumed, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		// End of headers
		if done {
			r.state = requestStateParsingBody
		}

		return n, nil
	case requestStateParsingBody:
		contentLength, err := strconv.Atoi(r.Headers.Get("content-length"))
		if err != nil {
			r.state = requestStateDone
			return 0, nil
		}

		r.Body = append(r.Body, data...)

		if len(r.Body) > contentLength {
			return 0, fmt.Errorf("body length %d exceeds content-length %d", len(r.Body), contentLength)
		}
		if len(r.Body) == contentLength {
			r.state = requestStateDone
		}

		return len(data), nil
	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("unknown state: %d", r.state)
	}
}

func parseRequestLine(bytes []byte) (*RequestLine, int, error) {
	str := string(bytes)
	crlfIndex := strings.Index(str, crlf)
	if crlfIndex == -1 {
		return nil, 0, nil
	}

	numBytesParsed := crlfIndex + len(crlf)
	parts := strings.Split(str[:crlfIndex], " ")

	if len(parts) != 3 {
		return nil, numBytesParsed, fmt.Errorf("invalid request line")
	}

	validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true, "HEAD": true, "OPTIONS": true}
	if !validMethods[parts[0]] {
		return nil, numBytesParsed, fmt.Errorf("invalid method: %s", parts[0])
	}

	if !strings.HasPrefix(parts[2], "HTTP/") {
		return nil, numBytesParsed, fmt.Errorf("invalid HTTP version: %s", parts[2])
	}

	version := strings.TrimPrefix(parts[2], "HTTP/")
	validVersions := map[string]bool{"1.1": true}
	if !validVersions[version] {
		return nil, numBytesParsed, fmt.Errorf("invalid HTTP version: %s", parts[2])
	}

	return &RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   version,
	}, numBytesParsed, nil
}
