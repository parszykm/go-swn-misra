package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run main.go <address> <port> <int>")
		return
	}

	address := os.Args[1]
	port := os.Args[2]
	numberStr := os.Args[3]

	// Convert the input string to a signed 32-bit integer
	number, err := strconv.ParseInt(numberStr, 10, 32)
	if err != nil {
		fmt.Println("Invalid integer:", err)
		return
	}

	// Create the address string
	serverAddress := fmt.Sprintf("%s:%s", address, port)

	// Connect to the server
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Printf("Failed to connect to %s: %s\n", serverAddress, err)
		return
	}
	defer conn.Close()

	// Set up channel to catch signals for graceful exit
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Main loop to send the integer every 10 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	fmt.Println("Sending integer every 10 seconds. Press CTRL+C to exit.")
	for {
		select {
		case <-ticker.C:
			err := binary.Write(conn, binary.BigEndian, int32(number))
			if err != nil {
				fmt.Println("Failed to send data:", err)
				return
			}
			fmt.Printf("Sent integer %d to %s\n", number, serverAddress)
		case <-signalChan:
			fmt.Println("\nReceived CTRL+C. Exiting...")
			return
		}
	}
}
