package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hailwind/udp-bench/config"
	kcp "github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
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
	conn.SetMtu(1350)
	defer conn.Close()
	smuxConfig := smux.DefaultConfig()
	smuxConfig.MaxReceiveBuffer = 4096 * 1024
	smuxConfig.KeepAliveInterval = time.Duration(10) * time.Second
	mux, err := smux.Server(conn, smuxConfig)
	if err != nil {
		log.Println(err)
		return
	}
	defer mux.Close()

	stream, err := mux.AcceptStream()
	if err != nil {
		log.Println(err)
		return
	}
	t2ichan := make(chan struct{})
	t2i := func() {
		defer close(t2ichan)
		frame := make([]byte, config.Mtu, config.Mtu)
		for {
			n, err := stream.Read([]byte(frame))
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
