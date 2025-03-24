package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func byte_reader(file *os.File) {
	for {
		input_b := make([]byte, 8)
		n, read_err := file.Read(input_b)
		if errors.Is(read_err, io.EOF) {
			break
		} else if read_err != nil {
			fmt.Printf("Error reading the file: %v\n", read_err.Error())
		}
		str := string(input_b[:n])
		fmt.Printf("read: %s\n", str)

	}
}

func line_reader(file *os.File) {
	var line string
	for {
		input_b := make([]byte, 8)
		n, read_err := file.Read(input_b)
		if read_err != nil {
			if line != "" {
				fmt.Printf("read: %s\n", line)
				line = ""
			}
			if errors.Is(read_err, io.EOF) {
				break
			}
			fmt.Printf("error: %s\n", read_err.Error())
			break
		}

		str := string(input_b[:n])
		parts := strings.Split(str, "\n")
		for i := 0; i < len(parts)-1; i++ {
			fmt.Printf("read: %s%s\n", line, parts[i])
			line = ""
		}
		line += parts[len(parts)-1]
	}
}
