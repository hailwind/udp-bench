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
	conn, err := net.Dial("udp", config.ServerAddr)
	checkError(err, "net.Dail")

	defer conn.Close()

	i2tchan := make(chan struct{})

	i2t := func() {
		defer close(i2tchan)
		frame := make([]byte, config.Mtu, config.Mtu)
		for {
			n, err := conn.Write([]byte(frame))
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
