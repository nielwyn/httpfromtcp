package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/response"
	"net/http"
	"strconv"
	"strings"
)

func handleProxy(w *response.Writer, path string) {
	resp, err := http.Get("https://httpbin.org/" + path)
	if err != nil {
		w.WriteStatusLine(response.StatusCodeInternalServerError)
		w.WriteHeaders(response.GetDefaultHeaders(0))
		w.WriteBody([]byte(err.Error()))
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusCode(resp.StatusCode))

	h := headers.Headers{}
	for k, v := range resp.Header {
		h[strings.ToLower(k)] = strings.Join(v, ", ")
	}
	h.Override("transfer-encoding", "chunked")
	h.Set("trailer", "X-Content-SHA256, X-Content-Length")
	h.Delete("content-length")
	w.WriteHeaders(h)

	var fullBody []byte
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			w.WriteChunkedBody(buf[:n])
			fullBody = append(fullBody, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	w.WriteChunkedBodyDone()

	trailers := headers.Headers{}
	for k, v := range resp.Trailer {
		trailers[strings.ToLower(k)] = strings.Join(v, ", ")
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(fullBody))
	trailers.Set("X-Content-SHA256", hash)
	trailers.Set("X-Content-Length", strconv.Itoa(len(fullBody)))
	w.WriteTrailers(trailers)
}
