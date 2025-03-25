package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	udp, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		fmt.Printf(" Error starting UDP: %v\n", err.Error())
		os.Exit(1)
	}

	udpConn, err := net.DialUDP("udp", nil, udp)
	if err != nil {
		fmt.Printf(" Error starting Connecting UDP: %v\n", err.Error())
		os.Exit(1)
	}
	defer udpConn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("> ")
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf(" Error reading user input: %v\n", err.Error())
		}

		_, err = udpConn.Write([]byte(str))
		if err != nil {
			fmt.Printf(" Error writing user input: %v\n", err.Error())
		}
	}
}
