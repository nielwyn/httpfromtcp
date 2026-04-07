package request

import (
	"fmt"
	"io"
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
	// Header      map[string]string
	// Body        []byte
	state requestState
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
			req.state = requestStateDone
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
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
		r.state = requestStateDone
		return consumed, nil
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
	requestLineParts := strings.Split(str[:crlfIndex], " ")

	if len(requestLineParts) != 3 {
		return nil, numBytesParsed, fmt.Errorf("invalid request line")
	}

	validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true, "HEAD": true, "OPTIONS": true}
	if !validMethods[requestLineParts[0]] {
		return nil, numBytesParsed, fmt.Errorf("invalid method: %s", requestLineParts[0])
	}

	if !strings.HasPrefix(requestLineParts[2], "HTTP/") {
		return nil, numBytesParsed, fmt.Errorf("invalid HTTP version: %s", requestLineParts[2])
	}

	version := strings.TrimPrefix(requestLineParts[2], "HTTP/")
	validVersions := map[string]bool{"1.1": true}
	if !validVersions[version] {
		return nil, numBytesParsed, fmt.Errorf("invalid HTTP version: %s", requestLineParts[2])
	}

	return &RequestLine{
		Method:        requestLineParts[0],
		RequestTarget: requestLineParts[1],
		HttpVersion:   version,
	}, numBytesParsed, nil
}
