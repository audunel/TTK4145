package main

import (
	"fmt"
	"net"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

type Counter struct {
	State int
}

type Message struct {
	Data string
}

func transmitServer(outChannel chan Message) {
	laddr, err := net.ResolveUDPAddr("udp", ":44556")
	if err != nil {
		log.Fatal(err)
	}

	baddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:33445")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		msg := <-outChannel
		_, err := conn.WriteToUDP([]byte(msg.Data), baddr)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func launchBackupProcess() {
	cmd := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run backup.go")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	outChannel := make(chan Message)
	go transmitServer(outChannel)

	launchBackupProcess()

	fmt.Println("Launching master process")
	counter := Counter{0}

	if len(os.Args) > 1 {
		initState, _ := strconv.Atoi(os.Args[1])
		counter.State = initState
		fmt.Printf("Master initiated with state %d\n", counter.State)
	}

	for {
		fmt.Printf("Current state: %d\n", counter.State)

		outChannel <- Message{strconv.Itoa(counter.State)}

		counter.State++
		time.Sleep(1 * time.Second)
	}
}
