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
	//conn, err := net.Dial("udp", config.ServerAddr)
	conn, err := kcp.DialWithOptions(config.ServerAddr, nil, 10, 3)
	conn.SetStreamMode(true)
	conn.SetMtu(1350)
	checkError(err, "net.Dail")

	defer conn.Close()

	smuxConfig := smux.DefaultConfig()
	smuxConfig.MaxReceiveBuffer = 4096 * 1024
	smuxConfig.KeepAliveInterval = time.Duration(10) * time.Second
	session, _ := smux.Client(conn, smuxConfig)
	stream, _ := session.OpenStream()

	i2tchan := make(chan struct{})

	i2t := func() {
		defer close(i2tchan)
		frame := make([]byte, config.Mtu, config.Mtu)
		for {
			n, err := stream.Write([]byte(frame))
			//fmt.Println("n:", n)
			checkError(err, "udp.Write", "n:", strconv.Itoa(n))
		}
	}

	go i2t()

	// wait for tunnel termination
	select {
	case <-i2tchan:
	}
}
