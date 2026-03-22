package common

import (
	"bytes"
	"fmt"
	"net"
	"sync"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

type Server struct {
	shutdownRequested bool
	clientConn        net.Conn
	serverListener    net.Listener
	mu                sync.Mutex
}

func NewServer(port int, listenBacklog int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	return &Server{
		serverListener: listener,
	}, nil
}

func (s *Server) Stop() {
	s.mu.Lock()
	s.shutdownRequested = true
	clientConn := s.clientConn
	s.clientConn = nil
	listener := s.serverListener
	s.serverListener = nil
	s.mu.Unlock()

	if clientConn != nil {
		_ = clientConn.Close()
		log.Info("action: close_client_socket | result: success")
	}

	if listener != nil {
		_ = listener.Close()
		log.Info("action: close_server_socket | result: success")
	}
}

func (s *Server) Run() {
	for !s.isShutdownRequested() {
		clientConn := s.acceptNewConnection()
		if clientConn == nil {
			continue
		}
		s.handleClientConnection(clientConn)
	}
}

func (s *Server) isShutdownRequested() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.shutdownRequested
}

func (s *Server) sendMessage(clientConn net.Conn, msgBytes []byte) error {
	totalSent := 0
	for totalSent < len(msgBytes) {
		sent, err := clientConn.Write(msgBytes[totalSent:])
		if err != nil {
			return err
		}
		if sent == 0 {
			return fmt.Errorf("socket closed before sending full message")
		}
		totalSent += sent
	}
	return nil
}

func (s *Server) handleClientConnection(clientConn net.Conn) {
	s.mu.Lock()
	s.clientConn = clientConn
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		currentConn := s.clientConn
		s.clientConn = nil
		s.mu.Unlock()

		if currentConn != nil {
			_ = currentConn.Close()
			log.Info("action: close_client_socket | result: success")
		}
	}()

	msgBytes := make([]byte, 1024)
	n, err := clientConn.Read(msgBytes)
	if err != nil {
		if !s.isShutdownRequested() {
			log.Errorf("action: receive_message | result: fail | error: %v", err)
		}
		return
	}
	if n == 0 {
		if !s.isShutdownRequested() {
			log.Errorf("action: receive_message | result: fail | error: %v", fmt.Errorf("socket closed before receiving message"))
		}
		return
	}

	msg := string(bytes.TrimRight(msgBytes[:n], " \t\r\n\v\f"))
	addr := clientConn.RemoteAddr().(*net.TCPAddr)
	log.Infof("action: receive_message | result: success | ip: %s | msg: %s", addr.IP.String(), msg)

	if err := s.sendMessage(clientConn, []byte(msg+"\n")); err != nil {
		if !s.isShutdownRequested() {
			log.Errorf("action: receive_message | result: fail | error: %v", err)
		}
	}
}

func (s *Server) acceptNewConnection() net.Conn {
	log.Info("action: accept_connections | result: in_progress")

	s.mu.Lock()
	listener := s.serverListener
	s.mu.Unlock()
	if listener == nil {
		return nil
	}

	clientConn, err := listener.Accept()
	if err != nil {
		if s.isShutdownRequested() {
			return nil
		}
		panic(err)
	}

	addr := clientConn.RemoteAddr().(*net.TCPAddr)
	log.Infof("action: accept_connections | result: success | ip: %s", addr.IP.String())
	return clientConn
}
