package common

import (
	// "bufio"
	// "fmt"
	// "net"
	// "os"
	// "os/signal"
	// "syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopLapse     time.Duration
	LoopPeriod    time.Duration
	PersonInfo    PersonData
}

// Client Entity that encapsulates how
type Client struct {
	config     ClientConfig
	lottery 	*LotteryConnection
	// conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:     config,
		lottery: nil,
	}
	return client
}

func (c *Client) PlayLottery() {
	lotteryConn, err := NewLotteryConnection(c.config.ServerAddress)
	if err != nil {
		log.Fatalf(
			"[CLIENT %v] Could not connect to server. Error: %v",
			c.config.ID,
			err,
		)
	}
	c.lottery = lotteryConn

	defer func() {
		c.lottery.Close() // TODO handle error
		c.lottery = nil
	}()

	c.submitDataToLottery()
	c.checkLotteryResult()

	// TODO expect response
}

func (c* Client) submitDataToLottery() {
	c.lottery.SendPersonInfo(c.config.PersonInfo)
	// TODO handle possible error
}

func (c* Client) checkLotteryResult() {
	result, err := c.lottery.GetResult()
	if err != nil {
		// TODO handle error
	}
	log.Infof("[CLIENT %v] The result of the lottery is: %v", c.config.ID, result)
}

// // StartClientLoop Send messages to the client until some time threshold is met
// func (c *Client) StartClientLoop() {
// 	// Create the connection the server in every loop iteration. Send an
// 	// autoincremental msgID to identify every message sent
// 	c.createClientSocket()
// 	msgID := 1

// 	sigtermSignal := make(chan os.Signal, 1) // TODO add this to the PlayLottery method
// 	signal.Notify(sigtermSignal, syscall.SIGTERM)

// loop:
// 	// Send messages if the loopLapse threshold has been not surpassed
// 	for timeout := time.After(c.config.LoopLapse); ; {
// 		select {
// 		case <-timeout:
// 			break loop
// 		case <-sigtermSignal:
// 			break loop
// 		default:
// 		}

// 		// Send
// 		fmt.Fprintf(
// 			c.conn,
// 			"[CLIENT %v] Message N°%v sent\n",
// 			c.config.ID,
// 			msgID,
// 		)
// 		msg, err := bufio.NewReader(c.conn).ReadString('\n')
// 		msgID++

// 		if err != nil {
// 			log.Errorf(
// 				"[CLIENT %v] Error reading from socket. %v.",
// 				c.config.ID,
// 				err,
// 			)
// 			c.conn.Close()
// 			return
// 		}
// 		log.Infof("[CLIENT %v] Message from server: %v", c.config.ID, msg)

// 		// Wait a time between sending one message and the next one
// 		time.Sleep(c.config.LoopPeriod)

// 		// Recreate connection to the server
// 		c.conn.Close()
// 		c.createClientSocket()
// 	}

// 	log.Infof("[CLIENT %v] Closing connection", c.config.ID)
// 	c.conn.Close()
// }
