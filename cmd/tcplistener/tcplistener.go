package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Println(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}

		go func(c net.Conn) {
			ch := getLinesChannel(c)
			for v := range ch {
				fmt.Println(v)
			}
			c.Close()
		}(conn)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	var lineBuffer strings.Builder
	ch := make(chan string)
	buf := make([]byte, 8)
	go func() {
		for {
			n, err := f.Read(buf)
			partsSlice := strings.Split(string(buf[:n]), "\n")

			for i, v := range partsSlice {
				lineBuffer.WriteString(v)
				if i == len(partsSlice)-1 {
					break
				}
				ch <- lineBuffer.String()
				lineBuffer.Reset()
			}

			if err == io.EOF {
				if lineBuffer.Len() > 0 {
					ch <- lineBuffer.String()
				}
				f.Close()
				close(ch)
				return
			}
		}
	}()
	return ch
}
