package main

import (
	"httpfromtcp/internal/response"
)

var body200 = []byte(`<html>
    <head>
      <title>200 OK</title>
    </head>
    <body>
      <h1>Success!</h1>
      <p>Your request was an absolute banger.</p>
    </body>
  </html>
  `)

var body400 = []byte(`<html>
    <head>
      <title>400 Bad Request</title>
    </head>
    <body>
      <h1>Bad Request</h1>
      <p>Your request honestly kinda sucked.</p>
    </body>
  </html>
  `)

var body500 = []byte(`<html>
    <head>
      <title>500 Internal Server Error</title>
    </head>
    <body>
      <h1>Internal Server Error</h1>
      <p>Okay, you know what? This one is on me.</p>
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

func handleYourProblem(w *response.Writer) {
	h := response.GetDefaultHeaders(len(body400))
	h.Override("content-type", "text/html")
	w.WriteStatusLine(response.StatusCodeBadRequest)
	w.WriteHeaders(h)
	w.WriteBody(body400)
}

func handleMyProblem(w *response.Writer) {
	h := response.GetDefaultHeaders(len(body500))
	h.Override("content-type", "text/html")
	w.WriteStatusLine(response.StatusCodeInternalServerError)
	w.WriteHeaders(h)
	w.WriteBody(body500)
}
