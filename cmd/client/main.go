package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	host string
	port int
)

func init() {
	flag.IntVar(&port, "port", 9000, "port to connect")
	flag.StringVar(&host, "host", "localhost", "remote host")
}
func main() {
	flag.Parse()

	address := fmt.Sprintf("%s:%d", host, port)
	log.Infof("Dial on address %s", address)
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		log.Errorf("Net.Dial err:%s", err)
	}
	log.Infof("Dial to address success")
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	var buf [128]byte

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Errorf("read content from stdin err:%s", err)
			os.Exit(1)
		}
		if _, err := conn.Write([]byte(line)); err != nil {
			log.Errorf("write message to conn err:%s", err)
			return
		}
		n, err := conn.Read(buf[:])
		if err != nil {
			log.Errorf("get message from server:%s", err)
			return
		}
		log.Infof("get message from server:%s", string(buf[:n]))
	}
}
