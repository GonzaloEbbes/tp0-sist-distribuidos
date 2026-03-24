package client

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/op/go-logging"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/internal/domain"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/internal/protocol"
)

const maxMessageSize = 8192
const REQUEST_TIMEOUT = 5 * time.Second
const winnersQueryRetryDelay = 200 * time.Millisecond

var log = logging.MustGetLogger("log")

type Client struct {
	serverAddress string
	clientID      string
	maxBatchSize  int
	maxBatchCount int
	conn          net.Conn
	encoder       *protocol.BetMessageEncoder
	decoder       *protocol.ServerResponseDecoder
	mu            sync.Mutex
}

func New(serverAddress string, clientID string, maxBatchCount int, encoder *protocol.BetMessageEncoder, decoder *protocol.ServerResponseDecoder) *Client {
	return &Client{
		serverAddress: serverAddress,
		clientID:      clientID,
		maxBatchSize:  maxMessageSize,
		maxBatchCount: maxBatchCount,
		encoder:       encoder,
		decoder:       decoder,
	}
}

func (c *Client) SendBet(bet domain.BetRequest) (domain.ServerResponse, error) {
	messages, err := c.encoder.EncodeBet(bet, c.maxBatchCount, c.maxBatchSize)
	if err != nil {
		return domain.ServerResponse{}, err
	}

	var lastResponse domain.ServerResponse
	for _, message := range messages {
		conn, err := net.DialTimeout("tcp", c.serverAddress, REQUEST_TIMEOUT)
		if err != nil {
			return domain.ServerResponse{}, err
		}
		if err := conn.SetDeadline(time.Now().Add(REQUEST_TIMEOUT)); err != nil {
			_ = conn.Close()
			return domain.ServerResponse{}, err
		}

		c.mu.Lock()
		c.conn = conn
		c.mu.Unlock()

		if err := writeAll(conn, message); err != nil {
			_ = c.Close()
			return domain.ServerResponse{}, err
		}

		responseBytes, err := readUntilDelimiter(conn, '\n', maxMessageSize)
		if err != nil {
			_ = c.Close()
			return domain.ServerResponse{}, err
		}

		lastResponse, err = c.decoder.DecodeResponse(responseBytes)
		_ = c.Close()
		if err != nil {
			return domain.ServerResponse{}, err
		}
		if !lastResponse.IsSuccess() {
			return lastResponse, nil
		}
	}

	return lastResponse, nil
}

func (c *Client) SendRequest(request domain.BetRequest) (domain.ServerResponse, error) {
	return c.SendBet(request)
}

func (c *Client) QueryWinnersUntilReady(request domain.BetRequest) (domain.ServerResponse, error) {
	for {
		response, err := c.SendRequest(request)
		if err != nil {
			return domain.ServerResponse{}, err
		}
		if !response.IsDrawNotReady() {
			return response, nil
		}
		time.Sleep(winnersQueryRetryDelay)
	}
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
