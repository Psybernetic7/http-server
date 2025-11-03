package main

import (
	"fmt"
	"log"
	"net"

	"github.com/Psybernetic7/http-server/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		listen, err := listener.Accept()
		if err != nil {
			continue
		}
		fmt.Println("Connection established")
		rq, err := request.RequestFromReader(listen)
		if err != nil {
			continue
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", rq.RequestLine.Method)
		fmt.Printf("- Target: %s\n", rq.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", rq.RequestLine.HttpVersion)
		fmt.Println("Connection Terminated")
	}

}
