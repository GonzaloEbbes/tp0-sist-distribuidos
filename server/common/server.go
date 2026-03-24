package common

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/op/go-logging"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/domain"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/server/internal/ports"
)

const maxMessageSize = 8192

var log = logging.MustGetLogger("log")

type Server struct {
	shutdownRequested bool
	clientConn        net.Conn
	serverListener    net.Listener
	decoder           ports.BetMessageDecoder
	encoder           ports.ResponseEncoder
	mu                sync.Mutex
}

func NewServer(port int, listenBacklog int, decoder ports.BetMessageDecoder, encoder ports.ResponseEncoder) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	return &Server{
		serverListener: listener,
		decoder:        decoder,
		encoder:        encoder,
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
	}

	if listener != nil {
		_ = listener.Close()
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
		}
	}()

	requestBytes, err := readUntilDelimiter(clientConn, '\n', maxMessageSize)
	if err != nil {
		if !s.isShutdownRequested() {
			errorCode := "malformed_message"
			if err.Error() == "message exceeds maximum size" {
				errorCode = "message_too_large"
			}
			s.respondWith(clientConn, domain.NewErrorResponse(errorCode, err.Error()))
		}
		return
	}

	s.respondWith(clientConn, s.decoder.Process(requestBytes))
}

func (s *Server) respondWith(clientConn net.Conn, response domain.Response) {
	responseBytes, err := s.encoder.Encode(response)
	if err != nil {
		return
	}

	_ = writeAll(clientConn, responseBytes)
}

func (s *Server) acceptNewConnection() net.Conn {
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
		return nil
	}

	return clientConn
}

func writeAll(conn net.Conn, message []byte) error {
	totalWritten := 0
	for totalWritten < len(message) {
		written, err := conn.Write(message[totalWritten:])
		if err != nil {
			return err
		}
		if written == 0 {
			return fmt.Errorf("socket closed before sending full message")
		}
		totalWritten += written
	}

	return nil
}

func readUntilDelimiter(conn net.Conn, delimiter byte, maxSize int) ([]byte, error) {
	buffer := make([]byte, 0, 256)
	chunk := make([]byte, 256)

	for {
		read, err := conn.Read(chunk)
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("socket closed before receiving full message")
			}
			return nil, err
		}
		if read == 0 {
			return nil, fmt.Errorf("socket closed before receiving full message")
		}

		buffer = append(buffer, chunk[:read]...)
		if len(buffer) > maxSize {
			return nil, fmt.Errorf("message exceeds maximum size")
		}

		if bytes.IndexByte(buffer, delimiter) >= 0 {
			messageEnd := bytes.IndexByte(buffer, delimiter) + 1
			return buffer[:messageEnd], nil
		}
	}
}
