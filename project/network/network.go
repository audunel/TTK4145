package network

import (
	"log"
	"net"
)

const (
	masterPort = "20021"
	slavePort  = "20022"
)

type IP string

type UDPMessage struct {
	Address IP
	Data    []byte
	Length  int
}

func GetOwnIP() IP {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Fatal(err)
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return IP(ipnet.IP.String())
				}
			}
		}
	}
	return "127.0.0.1"
}

func UDPInit(master bool, sendChannel, receiveChannel chan UDPMessage, networkLogger log.Logger) {
	var localPort, broadcastPort string
	if master {
		networkLogger.Print("Connecting as master")
		localPort = masterPort
		broadcastPort = slavePort
	} else {
		networkLogger.Print("Connecting as slave")
		localPort = slavePort
		broadcastPort = masterPort
	}

	laddr, err := net.ResolveUDPAddr("udp", ":"+localPort)
	if err != nil {
		networkLogger.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		networkLogger.Println("Failed to connect")
	}
	defer conn.Close()

	go listenServer(conn, receiveChannel, networkLogger)
	broadcastServer(conn, broadcastPort, sendChannel, networkLogger)
}

func listenServer(conn *net.UDPConn, receiveChannel chan UDPMessage, networkLogger log.Logger) {
	networkLogger.Printf("Listening on %s", conn.LocalAddr().String())
	for {
		buf := make([]byte, 1024)
		len, raddr, _ := conn.ReadFromUDP(buf)
		receiveChannel <- UDPMessage{Address: IP(raddr.IP.String()), Data: buf[:len], Length: len}
	}
}

func broadcastServer(conn *net.UDPConn, port string, sendChannel chan UDPMessage, networkLogger log.Logger) {
	networkLogger.Printf("Broadcasting to port %s", port)
	baddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+port)
	if err != nil {
		networkLogger.Fatal(err)
	}

	for {
		message := <-sendChannel
		conn.WriteToUDP(message.Data, baddr)
	}
}
