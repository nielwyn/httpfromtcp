package main

import (
	"httpfromtcp/internal/response"
	"os"
)

func handleVideo(w *response.Writer) {
	data, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		w.WriteStatusLine(response.StatusCodeInternalServerError)
		w.WriteHeaders(response.GetDefaultHeaders(0))
		w.WriteBody([]byte(err.Error()))
		return
	}
	h := response.GetDefaultHeaders(len(data))
	h.Override("content-type", "video/mp4")
	w.WriteStatusLine(response.StatusCodeOK)
	w.WriteHeaders(h)
	w.WriteBody(data)
}
