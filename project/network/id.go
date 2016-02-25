package network

import (
	"net"
	"log"
)

type ID string

func GetSenderID(sender *net.UDPAddr) ID {
	return ID(sender.IP.String())
}

func GetOwnID() ID {
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
					return ID(ipnet.IP.String())
				}
			}
		}
	}
	return "127.0.0.1"
}
