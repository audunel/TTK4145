// TODO: Make unit test using go testing framework
package main

import (
	"../network"
	"fmt"
	"time"
)

var localIP = network.GetOwnIP()

func printUDPMessage(msg network.UDPMessage) {
	fmt.Printf("msg: \n\t raddr = %s \n\t data = %s \n\t length = %v\n", msg.Address, string(msg.Data), msg.Length)
}

func main() {
	sendCh := make(chan network.UDPMessage)
	receiveCh := make(chan network.UDPMessage)

	go network.UDPInit("20001", "20001", sendCh, receiveCh)

	for {
		time.Sleep(1 * time.Second)

		msgText := "Hello me!"
		msg := network.UDPMessage{Address: localIP, Data: []byte(msgText), Length:len(msgText)}
		fmt.Println("Sending------")
		sendCh <- msg
		printUDPMessage(msg)
		fmt.Println("Receiving----")
		rcvMsg := <- receiveCh
		printUDPMessage(rcvMsg)
	}
}
