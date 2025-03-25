package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func getLinesChannel(conn net.Conn) <-chan string {
	lines := make(chan string)
	go func() {
		defer conn.Close()
		defer close(lines)
		currentLineContents := ""
		for {
			input_b := make([]byte, 8)
			n, err := conn.Read(input_b)
			if err != nil {
				if currentLineContents != "" {
					lines <- currentLineContents
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("Error reading data: %v", err.Error())
				return
			}
			string := string(input_b[:n])
			parts := strings.Split(string, "\n")
			for i := 0; i < len(parts)-1; i++ {
				lines <- fmt.Sprintf("%s%s", currentLineContents, parts[i])
				currentLineContents = ""
			}
			currentLineContents += parts[len(parts)-1]
		}
	}()
	return lines
}

func lineChannelFileReader(inputFilePath string) {
	file, err := os.Open(inputFilePath)
	if err != nil {
		fmt.Printf("Error opening the file: %v", err.Error())
		return
	}

	fmt.Printf("READING DATA FROM - %s\n", inputFilePath)
	fmt.Println("====================================")

	linesChan := getLinesChannel_file(file)
	for line := range linesChan {
		fmt.Printf("read: %s\n", line)
	}
}

// This will be run in go-routine
func getLinesChannel_file(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer f.Close()
		defer close(lines)
		currentLineContents := ""
		for {
			input_b := make([]byte, 8)
			n, err := f.Read(input_b)
			if err != nil {
				if currentLineContents != "" {
					lines <- currentLineContents
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("Error reading data: %v", err.Error())
				return
			}
			string := string(input_b[:n])
			parts := strings.Split(string, "\n")
			for i := 0; i < len(parts)-1; i++ {
				lines <- fmt.Sprintf("%s%s", currentLineContents, parts[i])
				currentLineContents = ""
			}
			currentLineContents += parts[len(parts)-1]
		}
	}()
	return lines
}
