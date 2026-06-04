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
	writerStateChunkedBody
	writerStateTrailers
	writerStateDone
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

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"content-length": strconv.Itoa(contentLen),
		"connection":     "close",
		"content-type":   "text/plain",
	}
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
	if w.writerState == writerStateStatusLine {
		return fmt.Errorf("cannot write headers: status line not written yet")
	}
	if w.writerState != writerStateHeaders {
		return fmt.Errorf("cannot write headers: already written")
	}

	for k, v := range headers {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}

	w.writer.Write([]byte("\r\n"))
	w.writerState = writerStateBody
	return nil
}

func (w *Writer) WriteBody(b []byte) (int, error) {
	if w.writerState == writerStateStatusLine {
		return 0, fmt.Errorf("cannot write body: status line not written yet")
	}
	if w.writerState == writerStateHeaders {
		return 0, fmt.Errorf("cannot write body: headers not written yet")
	}
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body: already written")
	}

	n, err := w.writer.Write(b)
	if err != nil {
		return n, err
	}
	w.writerState = writerStateDone
	return n, nil
}

func (w *Writer) WriteChunkedBody(b []byte) (int, error) {
	if w.writerState != writerStateBody && w.writerState != writerStateChunkedBody {
		return 0, fmt.Errorf("cannot write chunked body: invalid state")
	}

	_, err := fmt.Fprintf(w.writer, "%x\r\n", len(b))
	if err != nil {
		return 0, err
	}
	n, err := w.writer.Write(b)
	if err != nil {
		return n, err
	}
	_, err = w.writer.Write([]byte("\r\n"))
	if err != nil {
		return n, err
	}
	w.writerState = writerStateChunkedBody
	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writerState != writerStateChunkedBody {
		return 0, fmt.Errorf("cannot write chunked body done: invalid state")
	}

	n, err := w.writer.Write([]byte("0\r\n"))
	if err != nil {
		return n, err
	}
	w.writerState = writerStateTrailers
	return n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.writerState != writerStateTrailers {
		return fmt.Errorf("cannot write trailers: invalid state")
	}

	for k, v := range h {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	w.writerState = writerStateDone
	return nil
}
