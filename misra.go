package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"time"
)

type Misra struct {
	m               int64
	ping            Token
	pong            Token
	globalPingState State
}

// global variables
//var m int64 = 0
//var ping Token
//var misra.pong Token

func (misra *Misra) regenerate(x int64) {
	misra.ping.SetValue(x)
	misra.pong.SetValue((-1) * misra.ping.GetValue())
}

func (misra *Misra) incarnate(x int64) {
	misra.ping.SetValue(x + 1)
	misra.pong.SetValue((-1) * misra.ping.GetValue())
}

func (misra *Misra) sendToken(token *Token, conn net.Conn) error {
	value := token.GetValue()
	err := binary.Write(conn, binary.BigEndian, value)
	if err != nil {
		return fmt.Errorf("write error: %v", err)
	}
	misra.m = value
	return nil
}

func (misra *Misra) ManageRecvToken(recvValue int64, nextNodeConn net.Conn) error {
	if recvValue > 0 { // PING received
		fmt.Println("Received PING: ", recvValue)
		misra.ping.SetValue(recvValue)
		if misra.m == recvValue {
			misra.regenerate(misra.ping.GetValue())
			fmt.Println("Lost PONG. Regenerated: ", misra.pong.GetValue())
			err := misra.sendToken(&misra.pong, nextNodeConn)
			if err != nil {
				return err
			}
		}
		go func() {
			randomSleep := rand.Intn(5) + 1
			fmt.Printf("Entering critical session for: %d seconds...", randomSleep)
			misra.globalPingState.ManageState(PingLock)
			time.Sleep(5 * time.Second)
			misra.globalPingState.ManageState(PingFree)
			fmt.Println("Leaving critical session...")
			err := misra.sendToken(&misra.ping, nextNodeConn)
			if err != nil {
				return
			}
			fmt.Println("Forwarded PING: ", misra.ping.GetValue())
		}()
	} else if recvValue < 0 { // PONG received
		fmt.Println("Received PONG: ", recvValue)
		misra.pong.SetValue(recvValue)
		if misra.globalPingState.GetState() == PingLock {
			misra.incarnate(misra.ping.GetValue())

		} else if misra.m == recvValue {
			misra.regenerate(misra.pong.GetValue())
			fmt.Println("Lost PING. Regenerated: ", misra.ping.GetValue())
			err := misra.sendToken(&misra.ping, nextNodeConn)
			if err != nil {
				return err
			}
		}
		time.Sleep(1 * time.Second)
		err := misra.sendToken(&misra.pong, nextNodeConn)
		if err != nil {
			return err
		}
		fmt.Println("Forwarded PONG: ", misra.pong.GetValue())
	}
	return nil
}
