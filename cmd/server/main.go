package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var (
	port int
	host string
)

func init() {
	flag.IntVar(&port, "port", 9000, "port to listern")
	flag.StringVar(&host, "host", "localhost", "host location")
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	var buffer []byte

	for {
		n, err := conn.Read(buffer)
		if err == io.EOF{
			return
		}
		if err != nil {
			log.Errorf("read from conn err:%s", err)
			return
		}
		log.Infof("Message length:%d",n)
		log.Infof("Message fron conn:%s", string(buffer))
	}
}

func main() {
	flag.Parse()
	address := fmt.Sprintf("%s:%d", host, port)
	log.Infof("listen on address %s",address)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Errorf("net.Listern err:%s", err)
	}
	defer listener.Close()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM)

	// ctx, cancel := context.WithCancel(context.Background())
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("listener accept err:%s", err)
			os.Exit(1)
		}
		log.Infof("accept a conn")
		go handleRequest(conn)
	}
}
