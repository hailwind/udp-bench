package main

import (
	"log"
	"net"
	"strconv"

	"github.com/hailwind/testing/config"
)

func checkError(err error, args ...string) {
	if err != nil {
		log.Println(err, args)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	conn, err := net.Dial("udp", config.ServerAddr)
	checkError(err, "net.Dail")

	defer conn.Close()

	i2tchan := make(chan struct{})

	i2t := func() {
		defer close(i2tchan)
		frame := make([]byte, 130, 1300)
		for {
			n, err := conn.Write([]byte(frame))
			checkError(err, "udp.Write", "n:", strconv.Itoa(n))
		}
	}

	go i2t()

	// wait for tunnel termination
	select {
	case <-i2tchan:
	}
}
