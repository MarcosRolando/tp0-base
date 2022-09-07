package common

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

type ClosedConnectionError struct {}

func (cce *ClosedConnectionError) Error() string {
	return fmt.Sprintf("connection was closed")
}

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopLapse     time.Duration
	LoopPeriod    time.Duration
	SleepTime			time.Duration
	MaxBatchSize	uint
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
		log.Fatalf("Failed to open dataset file")
	}
	defer dataReader.Close()

	lotteryConn, err := NewLotteryConnection(c.config.ServerAddress)
	if err != nil {
		log.Fatalf(
			"Could not connect to server. Error: %v",
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
			log.Infof("Received SIGTERM signal, terminating client...")
			c.lottery.Close()
		}
	}()

	totalParticipants := 0
	totalWinners := 0
	for {
		batchInfo, err := dataReader.ReadNextBatch(c.config.MaxBatchSize)
		if err != nil {
			log.Panicf("Failed to read batch data. Error: %v", err.Error())
		}
		if len(batchInfo) == 0 { break }
		totalParticipants += len(batchInfo)
		if err := c.submitDataToLottery(batchInfo); err != nil {
			return; // Connection was closed by sigterm signal
		}
		batchWinners, err := c.checkLotteryResult()
		if err != nil { return } // Connection was closed by sigterm signal
		totalWinners += len(batchWinners)
	}

	if err := c.lottery.NotifyCompletion(); err != nil {
		if c.lottery.ClosedConn { return }
		log.Panicf("Failed to notify data completion to Lotter. Error: %v", err.Error())
	}
	log.Infof("Sent all data to Lottery. Winners rate: %v", float64(totalWinners) / float64(totalParticipants))
	
	received_result := false
	for !received_result {
		rType, val, err := c.lottery.GetFinalResult()
		if err != nil {
			if c.lottery.ClosedConn { return }
			log.Panicf("Failed to fetch final Lottery result. Error: %v", err.Error())
		}

		switch rType {
		case RemainingAgencies:
			log.Infof("Still processing %v agencies", val) 
			time.Sleep(c.config.SleepTime)
		case TotalWinners:
			log.Infof("The total amount of winners is %v", val) 
			received_result = true
		}
	}
}

func (c *Client) submitDataToLottery(batchInfo []PersonData) error {
	if err := c.lottery.SendBatchInfo(batchInfo); err != nil {
		if c.lottery.ClosedConn { return &ClosedConnectionError{} }
		log.Panicf("Failed to submit data to Lottery. Error: %v", err.Error())
	}
	log.Infof("Submitted batch data to Lottery")
	return nil
}

func (c *Client) checkLotteryResult() ([]PersonData, error) {
	winners, err := c.lottery.GetBatchResult()
	if err != nil {
		if c.lottery.ClosedConn { return nil, &ClosedConnectionError{} }
		log.Panicf("Failed to get result from Lottery. Error: %v", err.Error())
	}
	log.Infof("Received batch winners")
	for _, w := range winners {
		log.Infof("\n\n Name: %v\n Surname: %v\n Document: %v\n Birthdate: %v\n", 
			w.Name, w.Surname, w.Document, w.Birthdate)
	}
	return winners, nil
}
