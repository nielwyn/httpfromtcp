package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	fmt.Println(addr)
	if err != nil {
		fmt.Println(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Println(err)
		}
	}
}
