package main

import (
	"log"
	"net"
	"os"
	"strconv"

	"github.com/hailwind/udp-bench/config"
)

func checkError(err error, args ...string) {
	if err != nil {
		log.Println(err, args)
	}
}

func main() {
	log.Println(os.Args)
	if len(os.Args) == 3 {
		config.ServerAddr = os.Args[1]
		config.Mtu, _ = strconv.Atoi(os.Args[2])
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	addr, err := net.ResolveUDPAddr("udp", config.ServerAddr)
	checkError(err, "resolveUDPAddr")

	conn, err := net.ListenUDP("udp", addr)
	checkError(err, "listenUDP")

	defer conn.Close()

	t2ichan := make(chan struct{})
	t2i := func() {
		defer close(t2ichan)
		frame := make([]byte, config.Mtu, config.Mtu)
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
