package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Println(err)
	}
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
		req.RequestLine.Method,
		req.RequestLine.RequestTarget,
		req.RequestLine.HttpVersion,
	)

}
