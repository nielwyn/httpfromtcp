package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
)

type StatusCode int

const (
	StatusCodeOK                  StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

type Writer struct {
	writer      io.Writer
	writerState writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writerStateStatusLine {
		return fmt.Errorf("cannot write status line: already written")
	}

	statusLine := map[StatusCode]string{
		200: "HTTP/1.1 200 OK",
		400: "HTTP/1.1 400 Bad Request",
		500: "HTTP/1.1 500 Internal Server Error",
	}
	_, err := fmt.Fprintf(w.writer, "%s\r\n", statusLine[statusCode])
	if err != nil {
		return err
	}

	w.writerState = writerStateHeaders
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != writerStateHeaders {
		return fmt.Errorf("cannot write headers: status line not written yet")
	}

	for k, v := range headers {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}

	w.writerState = writerStateBody
	return nil
}

func (w *Writer) WriteBody(b []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body: headers not written yet")
	}

	w.writer.Write([]byte("\r\n"))
	return w.writer.Write(b)
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"content-length": strconv.Itoa(contentLen),
		"connection":     "close",
		"content-type":   "text/plain",
	}
}
