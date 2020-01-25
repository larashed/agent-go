package server

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
)

type DataHandler func(string)

type DomainSocketServer interface {
	Start(handler DataHandler) error
}

type Server struct {
	address  string
	listener net.Listener
}

func NewServer(address string) *Server {
	return &Server{address: address}
}

func (s *Server) Start(handler DataHandler) error {
	err := syscall.Unlink(s.address)
	if err != nil {
		// we can ignore this
	}

	listener, err := net.Listen("unix", s.address)
	if err != nil {
		return err
	}

	defer listener.Close()

	s.listener = listener

	channel := make(chan os.Signal, 1)
	signal.Notify(channel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGKILL,
	)

	go func(signal chan os.Signal) {
		<-channel
		s.Stop()
	}(channel)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return errors.Wrap(err, "Input error from connection")
		}

		go s.handleData(conn, handler)
	}
}

func (s *Server) handleData(c net.Conn, handler DataHandler) {
	for {
		buf := make([]byte, 5000)
		nr, err := c.Read(buf)
		if err != nil {
			return
		}

		data := buf[0:nr]
		log.Printf("Received message: %s", data)

		go handler(string(data))
	}
}

func (s *Server) Stop() error {
	err := s.listener.Close()
	if err != nil {
		return errors.Wrapf(err, "Failed to close listener to socket %v", s.address)
	}

	return nil
}
