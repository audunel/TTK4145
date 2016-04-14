package network

import (
	"net"
	"log"
)

type IP string

type UDPMessage struct {
	Address IP
	Data	[]byte
	Length	int
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

func UDPInit(localPort, broadcastPort string, sendChannel, receiveChannel chan UDPMessage, logger log.Logger) {
	laddr, err := net.ResolveUDPAddr("udp", ":"+localPort)
	if err != nil {
		logger.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		logger.Fatal(err)
	}
	defer conn.Close()

	go listenServer(conn, receiveChannel, logger)
	broadcastServer(conn, broadcastPort, sendChannel, logger)
}

func listenServer(conn *net.UDPConn, receiveChannel chan UDPMessage, logger log.Logger) {
	for {
		buf := make([]byte, 1024)
		len, raddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatal(err)
		}
		receiveChannel <- UDPMessage{Address: IP(raddr.IP.String()), Data: buf[:len], Length: len}
	}
}

func broadcastServer(conn *net.UDPConn, port string, sendChannel chan UDPMessage, logger log.Logger) {
	baddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+port)
	if err != nil {
		logger.Fatal(err)
	}

	for {
		message := <- sendChannel
		_, err := conn.WriteToUDP(message.Data, baddr)
		if err != nil {
			logger.Fatal(err)
		}
	}
}
