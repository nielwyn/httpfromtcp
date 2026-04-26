package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

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

func main() {
	server, err := server.Serve(
		port,
		func(w *response.Writer, req *request.Request) {
			if req.RequestLine.RequestTarget == "/yourproblem" {
				w.WriteStatusLine(response.StatusCodeBadRequest)
				h := response.GetDefaultHeaders(len(body400))
				h.Override("content-type", "text/html")
				w.WriteHeaders(h)
				w.WriteBody(body400)
				return
			}
			if req.RequestLine.RequestTarget == "/myproblem" {
				w.WriteStatusLine(response.StatusCodeInternalServerError)
				h := response.GetDefaultHeaders(len(body500))
				h.Override("content-type", "text/html")
				w.WriteHeaders(h)
				w.WriteBody(body500)
				return
			}
			if req.RequestLine.RequestTarget == "/" {
				w.WriteStatusLine(response.StatusCodeOK)
				h := response.GetDefaultHeaders(len(body200))
				h.Override("content-type", "text/html")
				w.WriteHeaders(h)
				w.WriteBody(body200)
				return
			}
		},
	)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
