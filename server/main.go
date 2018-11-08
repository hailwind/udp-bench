package main

import (
	"log"
	"net"

	"github.com/hailwind/udp-bench/config"
)

func checkError(err error, args ...string) {
	if err != nil {
		log.Println(err, args)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	addr, err := net.ResolveUDPAddr("udp", config.ServerAddr)
	checkError(err, "resolveUDPAddr")

	conn, err := net.ListenUDP("udp", addr)
	checkError(err, "listenUDP")

	defer conn.Close()

	t2ichan := make(chan struct{})
	t2i := func() {
		defer close(t2ichan)
		frame := make([]byte, 9600, 9600)
		for {
			conn.ReadFromUDP([]byte(frame))
		}
	}
	go t2i()
	// wait for tunnel termination
	select {
	case <-t2ichan:
	}
}
