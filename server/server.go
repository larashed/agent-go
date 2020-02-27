package server

import (
	"bufio"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"

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
	if fileExists(s.address) {
		if err := syscall.Unlink(s.address); err != nil {
			log.Err(err).Msg("failed to delete existing socket")
		}
	}

	s.listener, err = net.Listen("unix", s.address)
	if err != nil {
		return errors.Wrapf(err, `socket "%s" opening failed`, s.address)
	}

	for {
		conn, err := s.listener.Accept()
		log.Debug().Msg("received a new connection")
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
	defer c.Close()
	var (
		buf = make([]byte, 1024)
		r   = bufio.NewReader(c)
	)

	var bts []byte

	for {
		n, err := r.Read(buf)
		if err != nil {
			break
		}
		bts = append(bts, buf[:n]...)
	}

	line := string(bts)
	log.Trace().Msgf("received message '%s' with length %d", line, len(line))

	handler(strings.TrimSpace(line))
}

func (s *Server) Stop() error {
	close(s.listenerStop)
	return s.listener.Close()
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
