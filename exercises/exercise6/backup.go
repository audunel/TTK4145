package main

import (
	"fmt"
	"time"
	"log"
	"net"
	"strconv"
	"os/exec"
)

type Counter struct {
	State int
}

type Message struct {
	Data string
}

func listenForMessages(inChannel chan Message) {
	laddr, err := net.ResolveUDPAddr("udp", ":33445")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatal(err)
		}
		inChannel <- Message{string(buf[:n])}
	}
}

func restartMaster(initCounter Counter) {
	arg := fmt.Sprintf("go run master.go %d", initCounter.State)
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", arg)
	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	inChannel := make(chan Message)
	go listenForMessages(inChannel)

	var primaryCounter Counter

	select {
	case msg := <- inChannel:
		primaryCounter.State, _ = strconv.Atoi(msg.Data)
		fmt.Printf("Received initial state %d\n", primaryCounter.State)
	case <-time.After(5 * time.Second):
		fmt.Println("No message received. Is primary online?")
	}

	for {
		select {
		case <-time.After(10 * time.Second):
			fmt.Println("No message received in 10 seconds. Restarting master...")
			restartMaster(primaryCounter)
		case msg := <-inChannel:
			primaryCounter.State, _ = strconv.Atoi(msg.Data)
			fmt.Printf("Update received. Counter state: %d\n", primaryCounter.State)
		}
	}
}
