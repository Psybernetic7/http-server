package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {

	chanString := make(chan string)
	currentLine := ""

	filebytes := make([]byte, 8)

	go func() {
		defer f.Close()
		defer close(chanString)
		for {
			n, err := f.Read(filebytes)
			if n > 0 {
				stringChunk := string(filebytes[:n])
				parts := strings.Split(stringChunk, "\n")

				for i := 0; i < len(parts)-1; i++ {
					line := currentLine + parts[i]
					chanString <- line
					currentLine = ""
				}
				currentLine += parts[len(parts)-1]
			}

			if err == io.EOF {
				if currentLine != "" {
					chanString <- currentLine
				}
				return
			}
			if err != nil {
				return
			}
		}
	}()
	return chanString
}

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
		out := getLinesChannel(listen)
		for line := range out {
			fmt.Println(line)
		}
		fmt.Println("Connection Terminated")
	}

}
