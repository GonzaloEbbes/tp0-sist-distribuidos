package client

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/op/go-logging"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/internal/domain"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/internal/protocol"
)

const maxMessageSize = 4096

var log = logging.MustGetLogger("log")

type Client struct {
	serverAddress string
	clientID      string
	conn          net.Conn
	encoder       *protocol.BetMessageEncoder
	decoder       *protocol.ServerResponseDecoder
	mu            sync.Mutex
}

func New(serverAddress string, clientID string, encoder *protocol.BetMessageEncoder, decoder *protocol.ServerResponseDecoder) *Client {
	return &Client{
		serverAddress: serverAddress,
		clientID:      clientID,
		encoder:       encoder,
		decoder:       decoder,
	}
}

func (c *Client) SendBet(bet domain.BetRequest) (domain.ServerResponse, error) {
	message, err := c.encoder.EncodeBet(bet)
	if err != nil {
		return domain.ServerResponse{}, err
	}

	conn, err := net.Dial("tcp", c.serverAddress)
	if err != nil {
		return domain.ServerResponse{}, err
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	defer c.Close()

	if err := writeAll(conn, message); err != nil {
		return domain.ServerResponse{}, err
	}

	responseBytes, err := readUntilDelimiter(conn, '\n', maxMessageSize)
	if err != nil {
		return domain.ServerResponse{}, err
	}

	return c.decoder.DecodeResponse(responseBytes)
}

func (c *Client) Close() error {
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()

	if conn == nil {
		return nil
	}

	if err := conn.Close(); err != nil {
		return err
	}

	log.Infof("action: close_socket | result: success | client_id: %v", c.clientID)
	return nil
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
