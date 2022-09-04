package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type LotteryResult string

const (
	Loser 	LotteryResult = "loser"
	Winner								= "winner"
)

type LotteryConnection struct {
	conn net.Conn
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

func (lc *LotteryConnection) SendPersonInfo(info PersonData) error {
	print("hola")
	encodedInfo := fmt.Sprintf("%s;%s;%s;%s", info.Name, info.Surname, 
														info.Document, info.Birthdate) // TODO encapsulate data serialization
	encodedInfoLength := uint16(len(encodedInfo))
	buffer := make([]byte, encodedInfoLength + 2) // + 2 for the length bytes
	binary.BigEndian.PutUint16(buffer, encodedInfoLength)
	copy(buffer[2:], []byte(encodedInfo))
	fmt.Printf("mystr:\t %v \n", buffer) // TODO volar esto
	_, err := lc.conn.Write(buffer) // TODO ver el n que devuelve
	return err
}

func (lc *LotteryConnection) GetResult() (LotteryResult, error) {
	buffer := make([]byte, 1)
	_, err := lc.conn.Read(buffer) // TODO check n
	if err != nil {
		return Loser, err
	}
	if buffer[0] != 0 && buffer[0] != 1 {
		// TODO handle error
	}
	resultMap := map[byte]LotteryResult{0: Loser, 1: Winner}
	return resultMap[buffer[0]], nil
}

func (lc *LotteryConnection) Close() error {
	return lc.conn.Close()
}
