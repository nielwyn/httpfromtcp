package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(
		port,
		func(w *response.Writer, req *request.Request) {
			requestTarget := req.RequestLine.RequestTarget

			after, found := strings.CutPrefix(requestTarget, "/httpbin/")
			if found {
				handleProxy(w, after)
				return
			}

			switch requestTarget {
			case "/video":
				handleVideo(w)
			case "/yourproblem":
				handleYourProblem(w)
			case "/myproblem":
				handleMyProblem(w)
			case "/":
				handleRoot(w)
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
