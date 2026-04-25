package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

const (
	StatusCodeOK                  = 200
	StatusCodeBadRequest          = 400
	StatusCodeInternalServerError = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := map[StatusCode]string{
		200: "HTTP/1.1 200 OK",
		400: "HTTP/1.1 400 Bad Request",
		500: "HTTP/1.1 500 Internal Server Error",
	}

	_, err := fmt.Fprintf(w, "%s\r\n", statusLine[statusCode])
	if err != nil {
		return err
	}
	return nil
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length": strconv.Itoa(contentLen),
		"Connection":     "close",
		"Content-Type":   "text/plain",
	}
}
