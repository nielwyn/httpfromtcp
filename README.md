# HTTP from TCP

I wanted to understand what actually happens when a browser sends a request. Not the framework level — the actual bytes. So I built an HTTP/1.1 server in Go using nothing but a raw TCP listener.

No `net/http`. Just sockets, buffers, and the RFC.

## What I built

A working HTTP server from scratch:

- **TCP listener** that accepts connections and spawns a goroutine per connection
- **HTTP request parser** that reads bytes off the wire incrementally — handles the case where a single `Read()` call gives you half a header, or a header and a half
- **Response writer** with a state machine that enforces the correct write order: status line → headers → body. If you try to write the body before the headers, it errors. Same idea as a real HTTP response writer, but I built the state machine myself
- **Chunked transfer encoding** — streams a response in chunks and appends trailers (SHA-256 hash + content length) when done
- **Header parsing** that follows the RFC — field names are validated character by character, duplicate headers get comma-joined, and keys are normalized to lowercase
- **Proxy handler** that forwards requests to httpbin.org and streams the response back using chunked encoding
- **Video endpoint** that serves an mp4 with the correct `Content-Type`

## The interesting parts

The request parser was the most fun to get right. HTTP comes in over a stream, so you can't just wait for the "whole request" — you have to handle whatever chunk the OS gives you and keep state between reads. I used a growing buffer and a simple state machine (`initialized → parsingHeaders → parsingBody → done`) to track where in the parse we are across multiple reads.

The chunked proxy handler was also interesting. Instead of buffering the entire upstream response, it streams chunks to the client as they arrive, then writes the trailers (including a SHA-256 of the full body) at the end.

## Running it

```bash
go run ./cmd/httpserver
```

Server listens on port `42069`. Hit it with:

```bash
curl -v http://localhost:42069/
curl -v http://localhost:42069/httpbin/get
curl -v http://localhost:42069/video --output out.mp4
```

## Running tests

```bash
go test ./...
```

## Stack

Go — no external dependencies beyond `testify` for assertions.
