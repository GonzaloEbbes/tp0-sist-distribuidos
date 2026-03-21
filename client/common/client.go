package common

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// closeConnection closes the current connection if it is still open.
func (c *Client) closeConnection() {
	if c.conn != nil {
		c.conn.Close()
		log.Infof("action: close_socket | result: success | client_id: %v", c.config.ID)
		c.conn = nil
	}
}

// StopClientLoop closes resources associated with the current client loop.
func (c *Client) StopClientLoop() {
	c.closeConnection()
}

func shouldStop(stopCh <-chan bool) bool {
	select {
	case <-stopCh:
		return true
	default:
		return false
	}
}

func (c *Client) sendMessage(msg string) error {
	messageBytes := []byte(msg)
	totalWritten := 0

	for totalWritten < len(messageBytes) {
		written, err := c.conn.Write(messageBytes[totalWritten:])
		if err != nil {
			return err
		}
		totalWritten += written
	}

	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(stopCh <-chan bool, doneCh chan<- bool) {
	defer func() {
		c.closeConnection()
		doneCh <- true
	}()

	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		if shouldStop(stopCh) {
			log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
			return
		}

		// Create the connection the server in every loop iteration. Send an
		c.createClientSocket()

		message := fmt.Sprintf(
			"[CLIENT %v] Message N°%v\n",
			c.config.ID,
			msgID,
		)
		if err := c.sendMessage(message); err != nil {
			c.closeConnection()
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		c.closeConnection()

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)

		// Wait a time between sending one message and the next one
		select {
		case <-stopCh:
			log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
			return
		case <-time.After(c.config.LoopPeriod):
		}

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
