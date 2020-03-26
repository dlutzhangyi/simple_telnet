package main

import (
	"context"
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
	flag.StringVar(&host, "host", "", "host location")
}

func handleRequest(conn net.Conn,ctx context.Context) {
	defer conn.Close()

	var buf [128]byte

	for {
		n, err := conn.Read(buf[:])
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Errorf("read from conn err:%s", err)
			return
		}
		log.Infof("get message:%s", string(buf[:n]))
		message := "hello" + string(buf[:n])
		if _, err := conn.Write([]byte(message)); err != nil {
			log.Errorf("write to conn err:%s", err)
		}
	}
}

func main() {
	flag.Parse()
	address := fmt.Sprintf("%s:%d", host, port)
	log.Infof("listen on address %s", address)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Errorf("net.Listern err:%s", err)
	}
	defer listener.Close()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM)
	ctx,cancel:=context.WithCancel(context.Background())

	go func(){
		<- sigCh
		cancel()
	}()

	// ctx, cancel := context.WithCancel(context.Background())
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("listener accept err:%s", err)
			os.Exit(1)
		}
		log.Infof("accept a conn")
		go handleRequest(conn,ctx)
	}
}
