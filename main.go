package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

var (
	port         = flag.String("port", "8083", "Port to run the server on")
	nextNodeAddr = flag.String("next", "", "Address of the next node")
	nextNodePort = flag.String("next-port", "", "Port of the next node")
	initFlag     = flag.Bool("init", false, "Initialize as the first node (true/false)")
)

func usage() {
	fmt.Printf("Usage: %s [options]\n", os.Args[0])
	flag.PrintDefaults()
}

func handleConn(conn net.Conn, wg *sync.WaitGroup, nextNodeAddr *string, nextNodePort *string, initConn net.Conn, misra *Misra) error {
	nextNodeInitialized := false
	var nextNodeConn net.Conn
	if initConn != nil {
		nextNodeConn = initConn
		nextNodeInitialized = true
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Printf("Error closing connection: %s\n", err)
		}
	}(conn)

	defer func() {
		if nextNodeConn != nil {
			err := nextNodeConn.Close()
			if err != nil {
				return
			}
			fmt.Println("Closed connection to next node")
		}
	}()

	defer wg.Done()

	for {
		var signedInt int64
		err := binary.Read(conn, binary.BigEndian, &signedInt)
		if err != nil {
			return fmt.Errorf("read error: %v\n", err)
		}

		if !nextNodeInitialized {
			fmt.Println("Initializing connection with a next node...")
			nextNodeAddrWithPort := fmt.Sprintf("%s:%s", *nextNodeAddr, *nextNodePort)
			nextNodeConn = retryConnection(nextNodeAddrWithPort)
			if nextNodeConn == nil {
				return fmt.Errorf("Failed to connect to next node")
			}

			nextNodeInitialized = true
		}

		go func() {
			err := misra.ManageRecvToken(signedInt, nextNodeConn)
			if err != nil {
				fmt.Printf("Error connecting to next node: %s\n", err)
				return
			}
		}()

	}
}

func retryConnection(nextNodeAddrWithPort string) net.Conn {
	const maxRetries = 5
	const retryDelay = time.Second * 2 // 2-second delay between retries

	var nextNodeConn net.Conn
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		nextNodeConn, err = net.Dial("tcp", nextNodeAddrWithPort)
		if err == nil {
			break
		}

		fmt.Printf("Attempt %d: Error connecting to next node: %s\n", attempt, err)
		time.Sleep(retryDelay)
	}

	if err != nil {
		fmt.Println("Failed to connect to next node after retries")
		return nil
	}
	return nextNodeConn
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *nextNodeAddr == "" || *nextNodePort == "" {
		fmt.Println("Error: next node address and port are required.")
		flag.Usage()
		os.Exit(1)
	}

	var nextNodeConn net.Conn

	var misra = Misra{}
	misra.ping.SetValue(1)
	misra.pong.SetValue(-1)

	portWithHost := fmt.Sprintf("0.0.0.0:%s", *port)
	init := *initFlag

	fmt.Printf("Starting server on port %s\n", portWithHost)
	ln, err := net.Listen("tcp", portWithHost)
	if err != nil {
		fmt.Println("Failed to start server:", err)
		os.Exit(1)
	}
	defer func(ln net.Listener) {
		err := ln.Close()
		if err != nil {
			fmt.Printf("Error closing listener: %s\n", err)
		}
	}(ln)

	if init {
		var err error
		fmt.Println("This is the init node...")
		fmt.Println("Initializing connection with a next node...")
		nextNodeAddrWithPort := fmt.Sprintf("%s:%s", *nextNodeAddr, *nextNodePort)
		nextNodeConn = retryConnection(nextNodeAddrWithPort)
		if nextNodeConn == nil {
			fmt.Println("Failed to connect to next node")
			return
		}

		fmt.Println("Sending first PING: ", misra.ping.GetValue())
		err = misra.sendToken(&misra.ping, nextNodeConn)
		if err != nil {
			return
		}

		fmt.Println("Sending first PONG: ", misra.pong.GetValue())
		err = misra.sendToken(&misra.pong, nextNodeConn)
		if err != nil {
			return
		}
	}

	wg := sync.WaitGroup{}

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("accepted connection")
	wg.Add(1)
	fmt.Println("Init: ", init)
	go func() {
		err := handleConn(conn, &wg, nextNodeAddr, nextNodePort, nextNodeConn, &misra)
		if err != nil {
			fmt.Printf("Error handling conn: %s\n", err)
		}
	}()
	wg.Wait()
}
