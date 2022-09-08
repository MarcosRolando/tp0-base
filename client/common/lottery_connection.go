package common

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

type FinalLotteryResult int

const (
	RemainingAgencies FinalLotteryResult = iota
	TotalWinners
)

type LotteryConnection struct {
	conn 				net.Conn
	ClosedConn	bool
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

func (lc *LotteryConnection) SendBatchInfo(batchInfo []PersonData) error {
	// Notify request type batch
	if _, err := lc.conn.Write([]byte{0}); err != nil {
		return err
	}
	// Notify data length
	if err := binary.Write(lc.conn, binary.BigEndian, uint16(len(batchInfo))); err != nil {
		return err
	}
	for _, personInfo := range batchInfo {
		if err := lc.sendPersonInfo(personInfo); err != nil {
			return err
		}
	}
	return nil
}

func (lc *LotteryConnection) sendPersonInfo(info PersonData) error {
	encodedInfo := fmt.Sprintf("%s;%s;%s;%s", info.Name, info.Surname,
		info.Document, info.Birthdate)
	encodedInfoLength := uint16(len(encodedInfo))
	buffer := make([]byte, encodedInfoLength+2) // + 2 for the length bytes
	binary.BigEndian.PutUint16(buffer, encodedInfoLength)
	copy(buffer[2:], []byte(encodedInfo))
	_, err := lc.conn.Write(buffer)
	return err
}

func (lc *LotteryConnection) GetBatchResult() ([]PersonData, error) {
	totalWinnersBuff := make([]byte, 2)
	if _, err := io.ReadFull(lc.conn, totalWinnersBuff); err != nil {
		return nil, err
	}
	totalWinners := binary.BigEndian.Uint16(totalWinnersBuff)
	winners := make([]PersonData, totalWinners)
	for i := 0; i < int(totalWinners); i++ {
		winner, err := lc.getWinner()
		if err != nil {
			return nil, err
		}
		winners[i] = winner
	}
	return winners, nil
}

func (lc *LotteryConnection) Close() error {
	lc.ClosedConn = true
	return lc.conn.Close()
}

func (lc *LotteryConnection) GetFinalResult() (FinalLotteryResult, int, int, error) {
	// Notify request type results
	if _, err := lc.conn.Write([]byte{1}); err != nil {
		return 0, 0, 0, err
	}
	resultTypeBuff := make([]byte, 1)
	if _, err := io.ReadFull(lc.conn, resultTypeBuff); err != nil {
		return 0, 0, 0, err
	}
	resultType := FinalLotteryResult(resultTypeBuff[0])

	switch resultType {

	case RemainingAgencies:
		agenciesBuff := make([]byte, 2)
		if _, err := io.ReadFull(lc.conn, agenciesBuff); err != nil {
			return 0, 0, 0, err
		}
		winnersCountBuff := make([]byte, 4)
		if _, err := io.ReadFull(lc.conn, winnersCountBuff); err != nil {
			return 0, 0, 0, err
		}
		return resultType, int(binary.BigEndian.Uint16(agenciesBuff)), 
			int(binary.BigEndian.Uint32(winnersCountBuff)), nil

	case TotalWinners:
		buff := make([]byte, 4)
		if _, err := io.ReadFull(lc.conn, buff); err != nil {
			return 0, 0, 0, err
		}
		return resultType, 0, int(binary.BigEndian.Uint32(buff)), nil
	}

	return 0, 0, 0, fmt.Errorf("bad protocol: received final result flag of value %v", resultTypeBuff[0])
}

func (lc *LotteryConnection) getWinner() (PersonData, error) {
	dataLenBuf := make([]byte, 2)
	if _, err := io.ReadFull(lc.conn, dataLenBuf); err != nil {
		return PersonData{}, err
	}
	dataLen := binary.BigEndian.Uint16(dataLenBuf)
	winnerDataBuf := make([]byte, dataLen)
	if _, err := io.ReadFull(lc.conn, winnerDataBuf); err != nil {
		return PersonData{}, err
	}
	winnerData := strings.Split(string(winnerDataBuf), ";")
	if len(winnerData) != 4 {
		return PersonData{}, errors.New("bad protocol, received invalid winner data")
	}
	return PersonData{
		Name:      winnerData[0],
		Surname:   winnerData[1],
		Document:  winnerData[2],
		Birthdate: winnerData[3],
	}, nil
}
