package server

import (
	"bufio"
	"net"
	"net/textproto"
	"syscall"

	"github.com/pkg/errors"
)

type DataHandler func(string)

type DomainSocketServer interface {
	Start(handler DataHandler) error
	Stop() error
}

type Server struct {
	address      string
	listener     net.Listener
	listenerStop chan struct{}
}

func NewServer(address string) *Server {
	return &Server{
		address:      address,
		listenerStop: make(chan struct{}),
	}
}

// ErrServerStopped is returned by the Server's Start() method after a call to Stop().
var ErrServerStopped = errors.New("server stopped")

// Start listens on the TCP network address in server.address and then
// calls given DataHandler to handle requests on incoming connections.
//
// Start always returns a non-nil error. After Stop(), the returned error is ErrServerStopped.
func (s *Server) Start(handler DataHandler) (err error) {
	// we can ignore this
	_ = syscall.Unlink(s.address)

	s.listener, err = net.Listen("unix", s.address)
	if err != nil {
		return errors.Wrapf(err, `socket "%s" opening failed`, s.address)
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.listenerStop:
				return ErrServerStopped
			default:
				return errors.Wrap(err, "Input error from connection")
			}
		}

		go s.handleData(conn, handler)
	}
}

func (s *Server) handleData(c net.Conn, handler DataHandler) {
	reader := bufio.NewReader(c)
	tp := textproto.NewReader(reader)

	defer c.Close()

	for {
		line, err := tp.ReadLine()
		if err != nil {
			break
		}

		go handler(line)
	}
}

func (s *Server) Stop() error {
	close(s.listenerStop)
	return s.listener.Close()
}
