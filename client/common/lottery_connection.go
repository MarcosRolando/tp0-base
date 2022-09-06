package common

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"io"
)

type LotteryResult string

const (
	Loser 	LotteryResult = "Better luck next time... Loser"
	Winner	LotteryResult	= "Winner winner chicken dinner, you won!"
)

type LotteryConnection struct {
	conn 			net.Conn
	sentInfo	bool
}

func NewLotteryConnection(address string) (*LotteryConnection, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &LotteryConnection{
		conn,
		false,
	}, nil
}

func (lc *LotteryConnection) SendPersonInfo(info PersonData) error {
	encodedInfo := fmt.Sprintf("%s;%s;%s;%s", info.Name, info.Surname, 
														info.Document, info.Birthdate)
	encodedInfoLength := uint16(len(encodedInfo))
	buffer := make([]byte, encodedInfoLength + 2) // + 2 for the length bytes
	binary.BigEndian.PutUint16(buffer, encodedInfoLength)
	copy(buffer[2:], []byte(encodedInfo))
	_, err := lc.conn.Write(buffer)
	if err == nil { lc.sentInfo = true }
	return err
}

func (lc *LotteryConnection) GetResult() (LotteryResult, error) {
	if !lc.sentInfo {
		return "", errors.New("person info was not previously sent to the Lottery")
	}
	buffer := make([]byte, 1)
	if _, err := io.ReadFull(lc.conn, buffer); err != nil {
		return "", err
	}
	resCode := buffer[0]
	isValidResponse := func() bool { return resCode == 0 || resCode == 1 };
	if !isValidResponse() {
		panic(fmt.Sprintf("Received an invalid lottery response status of %v", resCode))
	}
	resultMap := map[byte]LotteryResult{0: Loser, 1: Winner}
	return resultMap[resCode], nil
}

func (lc *LotteryConnection) Close() error {
	return lc.conn.Close()
}
