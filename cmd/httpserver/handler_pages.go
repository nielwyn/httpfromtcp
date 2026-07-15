package main

import (
	"httpfromtcp/internal/response"
)

var body200 = []byte(`<html>
    <head>
      <title>200 OK</title>
    </head>
    <body>
      <h1>Success</h1>
      <p>The server parsed and handled this request successfully.</p>
    </body>
  </html>
  `)

var body400 = []byte(`<html>
    <head>
      <title>400 Bad Request</title>
    </head>
    <body>
      <h1>Bad Request</h1>
      <p>The request could not be understood by the server.</p>
    </body>
  </html>
  `)

var body500 = []byte(`<html>
    <head>
      <title>500 Internal Server Error</title>
    </head>
    <body>
      <h1>Internal Server Error</h1>
      <p>The server encountered an unexpected condition.</p>
    </body>
  </html>
  `)

func handleRoot(w *response.Writer) {
	h := response.GetDefaultHeaders(len(body200))
	h.Override("content-type", "text/html")
	w.WriteStatusLine(response.StatusCodeOK)
	w.WriteHeaders(h)
	w.WriteBody(body200)
}

func handleBadRequest(w *response.Writer) {
	h := response.GetDefaultHeaders(len(body400))
	h.Override("content-type", "text/html")
	w.WriteStatusLine(response.StatusCodeBadRequest)
	w.WriteHeaders(h)
	w.WriteBody(body400)
}

func handleServerError(w *response.Writer) {
	h := response.GetDefaultHeaders(len(body500))
	h.Override("content-type", "text/html")
	w.WriteStatusLine(response.StatusCodeInternalServerError)
	w.WriteHeaders(h)
	w.WriteBody(body500)
}
