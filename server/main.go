package main

import (
	"log"
	"os"
	"strconv"

	"github.com/hailwind/udp-bench/config"
	kcp "github.com/xtaci/kcp-go"
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
	lis, _ := kcp.ListenWithOptions(config.ServerAddr, nil, 10, 3)
	conn, _ := lis.AcceptKCP()
	conn.SetStreamMode(true)
	defer conn.Close()

	t2ichan := make(chan struct{})
	t2i := func() {
		defer close(t2ichan)
		frame := make([]byte, config.Mtu, config.Mtu)
		for {
			n, err := conn.Read([]byte(frame))
			checkError(err, "udp.Read", "n:", strconv.Itoa(n))
			//fmt.Println(n)
			//conn.ReadFromUDP([]byte(frame))
		}
	}
	go t2i()
	// wait for tunnel termination
	select {
	case <-t2ichan:
	}
}
