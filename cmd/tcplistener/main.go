package main

import (
	"fmt"
	"net"
	"os"
)

const inputFilePath = "messages.txt"

func main() {
	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Println("Error setting up listener: ", err.Error())
		os.Exit(1)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error setting up connection: ", err.Error())
			os.Exit(1)

		} else {
			fmt.Println("~CONNECTION ACCEPTED~")
		}
		lines := getLinesChannel(conn)

		for line := range lines {
			fmt.Printf("%s\n", line)
		}
		fmt.Println("Channel is closed...")
	}
}
