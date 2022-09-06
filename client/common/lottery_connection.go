package common

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

type LotteryConnection struct {
	conn     net.Conn
}

func NewLotteryConnection(address string) (*LotteryConnection, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &LotteryConnection{
		conn,
	}, nil
}

func (lc *LotteryConnection) SendBatchInfo(batchInfo []PersonData) error {
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
		if err != nil { return nil, err }
		winners[i] = winner
	}
	return winners, nil
}

func (lc *LotteryConnection) Close() error {
	return lc.conn.Close()
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
		Name: winnerData[0],
		Surname: winnerData[1],
		Document: winnerData[2],
		Birthdate: winnerData[3],
	}, nil
}
