# httpfromtcp

An HTTP/1.1 server implemented in Go directly on top of a raw TCP listener. It does not use `net/http` — request parsing, response writing, and connection handling are all built from the socket up, following RFC 9112.

## Features

- **Connection handling** — a TCP listener accepts connections and serves each one on its own goroutine
- **Incremental request parser** — reads bytes off the wire as they arrive and keeps parse state across reads, so a single `Read()` returning half a header (or a header and a half) is handled correctly
- **Stateful response writer** — enforces the correct write order (status line → headers → body) via a state machine; out-of-order writes return an error
- **Chunked transfer encoding** — streams response bodies in chunks and appends trailers (SHA-256 hash and content length) after the final chunk
- **RFC-compliant header parsing** — field names are validated character by character, duplicate headers are comma-joined, and keys are normalized to lowercase
- **Reverse proxy handler** — forwards requests to httpbin.org and streams the upstream response back with chunked encoding, without buffering the full body
- **Static video handler** — serves an mp4 file with the correct `Content-Type`

## Design notes

The request parser treats HTTP as what it is: a byte stream. Rather than waiting for a "complete request", it consumes whatever the OS delivers on each read, accumulating into a growing buffer and advancing a state machine (`initialized → parsingHeaders → parsingBody → done`) until the request is fully parsed.

The proxy handler streams end-to-end: chunks are forwarded to the client as they arrive from upstream, and trailers — including a SHA-256 of the complete body — are written once the upstream response ends.

## Project layout

```
cmd/httpserver/    HTTP server binary and route handlers
cmd/tcplistener/   Raw TCP listener utility
cmd/udpsender/     UDP sender utility
internal/request/  Incremental HTTP request parser
internal/response/ Response writer and chunked encoding
internal/headers/  RFC-compliant header parsing
internal/server/   TCP accept loop and connection lifecycle
```

## Usage

```bash
go run ./cmd/httpserver
```

The server listens on port `42069`:

```bash
curl -v http://localhost:42069/
curl -v http://localhost:42069/badrequest
curl -v http://localhost:42069/error
curl -v http://localhost:42069/httpbin/get
curl -v http://localhost:42069/video --output out.mp4
```

## Tests

```bash
go test ./...
```

## Dependencies

None — standard library only.
