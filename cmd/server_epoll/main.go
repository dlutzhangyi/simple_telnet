package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const (
	MaxEvents = 32
)

var (
	port int
	host string
)

func init() {
	flag.IntVar(&port, "port", 9000, "port to listen")
	flag.StringVar(&host, "host", "", "host address")
}

func handleRequest(fd int) {
	defer syscall.Close(fd)
	var buf [256]byte
	for {
		n, err := syscall.Read(fd, buf[:])
		if err != nil {
			log.Errorf("read from fd [%d] err:%s", err)
			return
		}
		if n > 0 {
			message := fmt.Sprintf("hello:%s", string(buf[:n]))
			if _, err := syscall.Write(fd, []byte(message)); err != nil {
				log.Errorf("write message [%s] to fd [%d] err:%s", message, fd, err)
				return
			}
		}
	}
}
func main() {
	var (
		event  syscall.EpollEvent
		events [MaxEvents]syscall.EpollEvent
	)
	// create socker
	listenFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		log.Errorf("create socket err:%s", err)
		os.Exit(1)
	}
	defer syscall.Close(listenFd)

	ipHost := net.ParseIP(host).To4()
	addr := syscall.SockaddrInet4{
		Port: port,
	}
	copy(addr.Addr[:], ipHost)

	//bind fd with addr
	if err := syscall.Bind(listenFd, &addr); err != nil {
		log.Errorf("bind socket err:%s", err)
		os.Exit(1)
	}
	//listen on fd, and set backlog 1024
	if err := syscall.Listen(listenFd, 1024); err != nil {
		log.Errorf("listen on fd err:%s", err)
		os.Exit(1)
	}

	//create epoll fd
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		log.Errorf("epoll create err:%s", err)
		os.Exit(1)
	}
	defer syscall.Close(epfd)

	event.Events = syscall.EPOLLIN | (syscall.EPOLLET & 0xffffffff)
	event.Fd = int32(listenFd)

	if err = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, listenFd, &event); err != nil {
		log.Errorf("epoll ctl err:%s", err)
		os.Exit(1)
	}
	for {
		n, err := syscall.EpollWait(epfd, events[:], -1)
		if err != nil {
			log.Errorf("epoll wait err:%s", err)
			break
		}
		log.Infof("epoll wait get events count:%d", n)
		for i := 0; i < n; i++ {
			if int(events[i].Fd) == listenFd {
				connFd, _, err := syscall.Accept(listenFd)
				if err != nil {
					log.Errorf("accept err:%s", err)
					continue
				}
				log.Infof("accept a conn, fd is %d", connFd)
				syscall.SetNonblock(listenFd, true)

				event.Events = syscall.EPOLLIN | (syscall.EPOLLET & 0xffffffff)
				event.Fd = int32(connFd)
				if err := syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, connFd, &event); err != nil {
					log.Errorf("epoll ctl err:%s", err)
					os.Exit(1)
				}
			} else {
				log.Infof("handle request with fd:%d", int(events[i].Fd))
				go handleRequest(int(events[i].Fd))
			}
		}
	}
}
