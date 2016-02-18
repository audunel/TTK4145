package network

import (
	"fmt"
	"net"
	"log"
	"strconv"
)

var laddr, baddr *net.UDPAddr

type UDPMessage struct {
	Raddr	string
	Data	string
	Length	int
}

func UDPInit(localListenPort, broadcastListenPort, msgSize int, sendCh, receiveCh chan UDPMessage) (err error) {
	baddr, err = net.ResolveUDPAddr("udp4", "255.255.255.255:"+strconv.Itoa(broadcastListenPort))
	if err != nil {
		return err
	}

	tempConn, err := net.DialUDP("udp4", nil, baddr)
	defer tempConn.Close()
	tempAddr := tempConn.LocalAddr()
	laddr, err := net.ResolveUDPAddr("udp4", tempAddr.String())
	laddr.Port = localListenPort

	fmt.Println(laddr)

	localListenConn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		return err
	}

	broadcastListenConn, err := net.ListenUDP("udp", baddr)
	if err != nil {
		localListenConn.Close()
		return err
	}

	go udpReceiveServer(localListenConn, broadcastListenConn, msgSize, receiveCh)
	go udpTransmitServer(localListenConn, broadcastListenConn, sendCh)

	return err
}

func udpTransmitServer(lconn, bconn *net.UDPConn, sendCh chan UDPMessage) {
	var err error

	for {
		msg := <-sendCh
		if msg.Raddr == "broadcast" {
			_, err = lconn.WriteToUDP([]byte(msg.Data), baddr)
		} else {
			raddr, err := net.ResolveUDPAddr("udp", msg.Raddr)
			if err != nil {
				log.Fatal(err)
				panic(err)
			}
			_, err = lconn.WriteToUDP([]byte(msg.Data), raddr)
		}
		if err != nil {
			log.Fatal(err)
			panic(err)
		}
	}
}

func udpReceiveServer(lconn, bconn *net.UDPConn, msgSize int, receiveCh chan UDPMessage) {
	bconnReceiveCh := make(chan UDPMessage)
	lconnReceiveCh := make(chan UDPMessage)

	go udpConnectionReader(lconn, msgSize, lconnReceiveCh)
	go udpConnectionReader(bconn, msgSize, bconnReceiveCh)

	for {
		select {
		case buf := <-bconnReceiveCh:
			receiveCh <- buf
		case buf := <-lconnReceiveCh:
			receiveCh <- buf
		}
	}
}

func udpConnectionReader(conn *net.UDPConn, msgSize int, receiveCh chan UDPMessage) {
	for {
		buf := make([]byte, msgSize)
		n, raddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatal(err)
			panic(err)
		}
		receiveCh <- UDPMessage{Raddr: raddr.String(), Data: string(buf[:n]), Length: n}
	}
}
