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

// DataHandler incoming data handler
type DataHandler func(string)

// Server holds socket server structure
type Server struct {
	socketType    string
	socketAddress string
	listener      net.Listener
	listenerStop  chan struct{}
}

// NewServer creates a new `Server` instance
func NewServer(networkType, networkAddress string) *Server {
	return &Server{
		socketType:    networkType,
		socketAddress: networkAddress,
		listenerStop:  make(chan struct{}),
	}
}

// ErrServerStopped is returned by the Server's Start() method after a call to Stop().
var ErrServerStopped = errors.New("server stopped")

// Start listens on the TCP network socketAddress in server.socketAddress and then
// calls given DataHandler to handle requests on incoming connections.
//
// Start always returns a non-nil error. After Stop(), the returned error is ErrServerStopped.
func (s *Server) Start(handler DataHandler) (err error) {
	// we can ignore this
	if s.socketType == "unix" && fileExists(s.socketAddress) {
		log.Debug().Msg("Socket exists. Trying to delete.")

		if err := syscall.Unlink(s.socketAddress); err != nil {
			log.Err(err).Msg("Failed to delete existing socket")
		}
	}

	s.listener, err = net.Listen(s.socketType, s.socketAddress)
	if err != nil {
		return errors.Wrapf(err, `Failed to open socket to "%s"`, s.socketAddress)
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

		log.Debug().Msg("Received a new connection")
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
	log.Trace().Msgf("Received message:\n'%s' with length %d", line, len(line))

	handler(strings.TrimSpace(line))
}

// Stop socket server
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
