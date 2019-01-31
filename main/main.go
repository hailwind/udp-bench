package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var buffSize = 65536
var host = flag.String("host", "0.0.0.0", "host")
var port = flag.String("p", "5301", "port")
var mode = flag.String("m", "server", "running mode, server or client")
var packetSize = flag.Int("ps", 128, "packet size")
var maxBandwidth = flag.Int("b", 50, "max bandwidth, Mbps")
var duration = flag.Int64("d", 10, "the all time of testing")
var isSleep = flag.String("s", "true", "to sleep or not to sleep")
var sleepTime = flag.Int64("st", 50, "the sleep time, micro seconds")

func checkError(err error, args ...string) {
	if err != nil {
		log.Println(err, args)
		os.Exit(1)
	}
}

func sleepMicroSecs() int {
	pps := *maxBandwidth * 1000 * 1000 / 8 / *packetSize
	return 1000000 / pps
}

func handleRecv(conn *net.UDPConn) (uint64, int) {
	buff := make([]byte, buffSize)
	n, _, err := conn.ReadFromUDP(buff)
	checkError(err)
	if n <= 0 {
		fmt.Println("read <= 0")
		return 0, 0
	}
	//fmt.Println("n: ", n)
	return binary.BigEndian.Uint64(buff), n
}

func recvAndSendResp(conn *net.UDPConn) string {
	buff := make([]byte, buffSize)
	n, rAddr, err := conn.ReadFromUDP(buff)
	checkError(err)
	if n > 0 {
		//time.Sleep(time.Duration(sleepTime) * time.Millisecond)
		n, err = conn.WriteToUDP(buff[0:n], rAddr)
		checkError(err)
		return string(buff[0:n])
	}
	return ""
}

func handleDial(conn *net.UDPConn) (int, int) {
	mbwStr := recvAndSendResp(conn)
	arr := strings.Split(mbwStr, "_")
	mbw, _ := strconv.Atoi(arr[1])

	rttStr := recvAndSendResp(conn)
	arr = strings.Split(rttStr, "_")
	rtt, _ := strconv.Atoi(arr[1])
	return mbw, rtt
}

func handleSend(conn *net.UDPConn, rtt int) {
	var lastSeq, recvCnt, recvBytes, missCnt, outOrderCnt uint64
	begin := time.Now().Unix()
	for {
		seq, len := handleRecv(conn)
		//fmt.Println("SEQ:", seq)
		if seq == 0 {
			fmt.Println("seq == 0")
			break
		}
		if seq == 0xffffffffffffffff {
			duration := time.Now().Unix() - begin
			avgBps := recvBytes * 8 / uint64(duration) / 1024 / 1024
			avgPps := recvCnt / uint64(duration)
			fmt.Println("============Report============")
			fmt.Println("Duration (seconds) : ", duration)
			fmt.Println("Bytes Recv (MBytes): ", recvBytes/1024/1024)
			fmt.Println("Packet Recv (Pkts): ", recvCnt)
			fmt.Println("Avg BandWidth (Mbps): ", avgBps)
			fmt.Println("Avg Packet Rate (pps): ", avgPps)
			fmt.Println("Fake Packet Lost: ", missCnt)
			fmt.Println("Real Packet Lost: ", lastSeq-recvCnt)
			fmt.Println("Out of Order: ", outOrderCnt)
			fmt.Println("Last sequence: ", lastSeq)
			fmt.Println("============Finish============")
			lastSeq = 0
			missCnt = 0
			break
		} else {
			recvCnt++
			recvBytes += uint64(len)
			if seq > lastSeq {
				diff := seq - lastSeq
				if diff == 1 {
					lastSeq = seq
				} else {
					fmt.Printf("Bigger seq: %d, lseq: %d, diff: %d\n", seq, lastSeq, diff)
					missCnt += diff
					lastSeq = seq
				}
			} else {
				diff := lastSeq - seq
				fmt.Printf("Smaller seq: %d, lseq: %d, diff: %d\n", seq, lastSeq, diff)
				outOrderCnt++
			}
		}
	}
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
	mbw, rtt := handleDial(conn)
	fmt.Println("RTT: ", rtt, "ms MBW: ", mbw, " Mbps")
	handleSend(conn, rtt)
}

func sendAndWaitRecv(conn *net.UDPConn, content string) string {
	n, err := conn.Write([]byte(content))
	checkError(err)
	if n > 0 {
		conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
		buff := make([]byte, buffSize)
		n, err = conn.Read(buff)
		checkError(err)
		return string(buff[0:n])
	}
	return ""
}

func dial(conn *net.UDPConn) {
	begin := time.Now().UnixNano()
	resp := sendAndWaitRecv(conn, fmt.Sprintf("MBW_%d", *maxBandwidth))
	fmt.Println("MBW:", resp)
	curr := time.Now().UnixNano()
	rtt := (curr - begin) / 1000 / 1000
	resp = sendAndWaitRecv(conn, fmt.Sprintf("RTT_%d", rtt))
	fmt.Println("RTT:", resp)
}

func randContent(seq uint64) *[]byte {
	buff := make([]byte, *packetSize+8)
	binary.BigEndian.PutUint64(buff, seq)
	// rand.Seed(time.Now().Unix())
	// for i := 8; i < *packetSize; i++ {
	// 	d := rand.Intn(26) + 64
	// 	buff[i] = byte(d)
	// }
	return &buff
}

func send(conn *net.UDPConn) {
	//sleepUsecs := sleepMicroSecs()
	//fmt.Println("Sleep micro seconds: ", sleepUsecs)
	begin := time.Now().Unix()
	var seq uint64
	for {
		seq++
		content := randContent(seq)
		//fmt.Println("SEQ:", seq)
		n, err := conn.Write(*content)
		checkError(err)
		if n < 0 {
			fmt.Println("Failed to send data")
			os.Exit(1)
		}
		if *isSleep == "True" || *isSleep == "true" || *isSleep == "T" || *isSleep == "t" {
			time.Sleep(time.Duration(*sleepTime) * time.Microsecond)
		}
		curr := time.Now().Unix()
		if (curr - begin) > *duration {
			content := randContent(0xffffffffffffffff)
			n, err := conn.Write(*content)
			checkError(err)
			if n < 0 {
				fmt.Println("Failed to send data")
				os.Exit(1)
			}
			break
		}
	}
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

	dial(conn)
	send(conn)
}

func main() {
	flag.Parse()

	if *mode == "server" || *mode == "s" {
		for {
			server()
		}
	} else if *mode == "client" || *mode == "c" {
		client()
	}
}
