package common

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
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
	config  ClientConfig
	lottery *LotteryConnection
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

func (c *Client) PlayLottery() {
	dataReader := DataReader{}
	if err := dataReader.Open(fmt.Sprintf("/datasets/dataset-%v.csv", c.config.ID)); err != nil {
		log.Fatalf("[CLIENT %v] Failed to open dataset file", c.config.ID)
	}
	defer dataReader.Close() // TODO see about handling error

	lotteryConn, err := NewLotteryConnection(c.config.ServerAddress)
	if err != nil {
		log.Fatalf(
			"[CLIENT %v] Could not connect to server. Error: %v",
			c.config.ID,
			err,
		)
	}
	c.lottery = lotteryConn
	defer c.lottery.Close() // Ignoring error as we are already closing the program, nothing to be done

	sigtermSignal := make(chan os.Signal, 1)
	signal.Notify(sigtermSignal, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigtermSignal)
		close(sigtermSignal)
	}()
	go func() {
		_, ok := <- sigtermSignal
		if ok {
			log.Infof("[CLIENT %v] Received SIGTERM signal, terminating client...", c.config.ID)
			c.lottery.Close() // TODO see about os.exit(0) maybe thread leak? race with panic maybe?
		}
	}()

	const maxBatchSize = 2000
	totalParticipants := 0
	totalWinners := 0
	for {
		batchInfo, err := dataReader.ReadNextBatch(maxBatchSize)
		if err != nil {
			log.Panicf("[CLIENT %v] Failed to read batch data. Error: %v", c.config.ID, err.Error())
		}
		if len(batchInfo) == 0 { break }
		totalParticipants += len(batchInfo)
		c.submitDataToLottery(batchInfo)
		batchWinners := c.checkLotteryResult()
		totalWinners += len(batchWinners)
	}

	if err := c.lottery.NotifyCompletion(); err != nil {
		log.Errorf("[CLIENT %v] Failed to notify data completion to Lotter. Error: %v", c.config.ID, err.Error())
	} else {
		log.Infof("[CLIENT %v] Sent all data to Lottery. Winners rate: %v", c.config.ID, 
			float64(totalWinners) / float64(totalParticipants))
	}
}

func (c *Client) submitDataToLottery(batchInfo []PersonData) {
	if err := c.lottery.SendBatchInfo(batchInfo); err != nil {
		log.Panicf("[CLIENT %v] Failed to submit data to Lottery. Error: %v", c.config.ID, err.Error())
	}
	log.Infof("[CLIENT %v] Submitted batch data to Lottery", c.config.ID)
}

func (c *Client) checkLotteryResult() []PersonData {
	winners, err := c.lottery.GetBatchResult()
	if err != nil {
		log.Panicf("[CLIENT %v] Failed to get result from Lottery. Error: %v", c.config.ID, err.Error())
	}
	log.Infof("[CLIENT %v] Received batch winners", c.config.ID)
	for _, w := range winners {
		log.Infof("\n\n Name: %v\n Surname: %v\n Document: %v\n Birthdate: %v\n", 
			w.Name, w.Surname, w.Document, w.Birthdate)
	}
	return winners
}
