package network

import (
	"net"
	"fmt"
	"log"
)

type tcpMessage struct {
	Raddr 	string
	Data	string
	Length	int
}

var connList map[string]*net.TCPConn
var connListMutex = &sync.Mutex{}
type tcpConn struct {
	conn	  *net.TCPConn
	recieveCh chan tcpMessage
}

func TCPInit(localListenPort int, sendCh, recieveCh chan tcpMessage) {
	fmt.Println("Initializing TCP")

	connList = make(map[string]*net.TCPConn)

	baddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:"+strconv.Itoa(20323)
	if err != nil {
		log.Fatal(err)
	}

	// Generates local address
	tempConn, err := net.DialUDP("udp4", nil, baddr)
	if err != nil {
		log.Fatal(err)
	}
	tempAddr := tempConn.LocalAddr()
	laddr, err := net.ResolveTCPAddr("tcp4", net.JoinHostPort(tempAddr.String(), localListenPort)
	if err != nil {
		log.Fatal(err)
	}
	tempConn.Close()

	listener, err := net.ListenTCP("tcp4", laddr)
	if err != nil {
		log.Fatal(err)
	}

	go TCPTransmitServer(sendCh)
	go AcceptConns(listener)
}

func TCPTransmitServer(ch chan TCPMessage) {
	for {
		msg := <-ch
		_, ok := connList[msg.Raddr]
		if ok != true {
			NewTCPConn(msg.Raddr)
		}
		connListMutex.Lock()
		sendConn, ok := connList[msg.Raddr]
		if(ok != true) {
			connListMutex.Unlock()
			err := errors.New("Failed to connect to " + msg.Raddr + "\n")
			panic(err)
		} else {
			n, err := sendConn.Write([]byte(msg.Data))
			connListMutex.Unlock()
			if err != nil || n < 0 {
				log.Fatal(err)
				connListMutex.Lock()
				delete(connList, msg.Raddr)
				conListMutex.Unlock()
			}
		}
	}
}

func AcceptConn(listener TCPListener, recieveCh chan tcpMessage) {
	// Listens for new connections
	for {
		newConn, err := listener.AcceptTCP()
		fmt.Println("Received new request for connection")
		if err != nil {
			log.Fatal(err)
		}
		raddr := newConn.RemoteAddr()
		connListMutex.Lock()
		connList[raddr.String()] = newConn
		connListMutex.Unlock()

		go ReadServer(raddr, newConn, receiveCh)
	}
}

func ReadServer(raddr string, conn *net.TCPConn, receiveCh chan tcpMessage) {
	// Reading server for connection conn
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil || n < 0 {
			log.Fatal(err)
			connListMutex.Lock()
			conn.Close()
			delete(connList, raddr)
			connListMutex.Unlock()
			return
		} else {
			receiveCh <- tcpMessage{Raddr:raddr, Data:string(buf[:n], Length:n}
		}
	}
}
