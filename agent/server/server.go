package server

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
)

type Server struct {
	address  string
	listener net.Listener
}

func NewServer(address string) *Server {
	return &Server{address: address}
}

type DataHandler func(string)

func (s *Server) Start(handler DataHandler) error {
	ln, err := net.Listen("unix", s.address)
	if err != nil {
		return err
	}

	log.Printf("Listening to %s", s.address)

	s.listener = ln

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
		log.Println("received signal")
		s.Stop()
	}(channel)

	for {
		conn, err := ln.Accept()
		if err != nil {
			return errors.Wrap(err, "Input error from connection")
		}

		go s.handleData(conn, handler)
	}
}

func (s *Server) handleData(c net.Conn, handler DataHandler) {
	for {
		buf := make([]byte, 512)
		nr, err := c.Read(buf)
		if err != nil {
			return
		}

		data := buf[0:nr]
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
