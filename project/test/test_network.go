// TODO: Make unit test using go testing framework
package main

import (
	"../network"
	"fmt"
	"time"
	"log"
)

var localIP = network.GetOwnID()

func printUDPMessage(msg network.UDPMessage) {
	fmt.Printf("msg: \n\t raddr = %s \n\t data = %s \n\t length = %v\n", msg.Raddr, string(msg.Data), msg.Length)
}

func main() {
	sendCh := make(chan network.UDPMessage)
	receiveCh := make(chan network.UDPMessage)

	err := network.UDPInit(20001, 20002, 1024, sendCh, receiveCh)
	if err != nil {
		log.Fatal(err)
	}

	for {
		time.Sleep(1 * time.Second)

		msg := network.UDPMessage{Raddr: string(localIP)+":20001", Data: []byte("Hello me!"), Length:9}
		fmt.Println("Sending------")
		sendCh <- msg
		printUDPMessage(msg)
		fmt.Println("Receiving----")
		rcvMsg := <- receiveCh
		printUDPMessage(rcvMsg)
	}
}
