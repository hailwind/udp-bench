package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

var buffSize = 1516
var host = flag.String("host", "0.0.0.0", "host")
var port = flag.String("port", "5301", "port")
var mode = flag.String("mode", "server", "running mode, server or client")
var packetSize = flag.Int("size", 64, "packet size")
var maxBandwidth = flag.Int("bandwidth", 50, "max bandwidth, Mbps")

func checkError(err error, args ...string) {
	if err != nil {
		log.Println(err, args)
	}
}

func sleepMicroSecs() int {
	pps := *maxBandwidth * 1000000 / 8 / *packetSize
	return 1000000 / pps
}

func handleRecv(conn *net.UDPConn) int {
	data := make([]byte, buffSize)
	n, _, err := conn.ReadFromUDP(data)
	if err != nil {
		fmt.Println("failed to read UDP msg because of ", err.Error())
		return -1
	}
	if n <= 0 {
		fmt.Println("read <= 0")
		return -1
	}
	var seq int
	buff := bytes.NewBuffer(data)
	binary.Read(buff, binary.BigEndian, &seq)
	return seq
}

func handleDial() {

}

func server() {
	addr, err := net.ResolveUDPAddr("udp", *host+":"+*port)
	if err != nil {
		fmt.Println("Can't resolve address: ", err)
		os.Exit(1)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()
	dialed := false
	lastSeq := 0
	missCnt := 0
	outOrder := 0
	for {
		if dialed {
			seq := handleRecv(conn)
			if seq == -1 {
				break
			}
			if seq == 0 {
				lastSeq = 0
				missCnt = 0
				fmt.Println("============Report============")
				fmt.Println("Packet Lost: ", missCnt)
				fmt.Println("Out of Order: ", outOrder)
				//fmt.Println("Last Sequence: ", lastSeq)
				fmt.Println("============Finish============")
			}
			if seq == (lastSeq + 1) {
				lastSeq++
			}
		} else {
			handleDial()
		}
	}
}

func dial() {

}

func waitDial() {

}

func send() {

}

func client() {
	addr, err := net.ResolveUDPAddr("udp", *host+":"+*port)
	if err != nil {
		fmt.Println("Can't resolve address: ", err)
		os.Exit(1)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	checkError(err)
	defer conn.Close()

	sleepUsecs := sleepMicroSecs()
	for {
		time.Sleep(time.Duration(sleepUsecs) * time.Microsecond)
	}
}

func main() {
	flag.Parse()

	if *mode == "server" {
		for {
			server()
		}
	} else {
		client()
	}
}
