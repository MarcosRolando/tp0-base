package common

import (
	"encoding/csv"
	"os"
)

type DataReader struct {
	file			*os.File
	reader 		*csv.Reader
}

func (dr *DataReader) Open(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil { return err }
	dr.file = file
	dr.reader = csv.NewReader(dr.file)
	dr.reader.FieldsPerRecord = 4 // name, surname, document, birthdate
	return nil
}

func (dr *DataReader) ReadNextBatch(batchSize uint) ([]PersonData, error) {
	buff := make([]PersonData, batchSize)
	for i := 0; i < int(batchSize); i++ {
		record, err := dr.reader.Read()
		if record == nil { return buff[:i], nil }
		if err != nil { return nil, err }
		buff[i] = PersonData{
			Name: record[0],
			Surname: record[1],
			Document: record[2],
			Birthdate: record[3],
		}
	}
	return buff, nil
}

func (dr *DataReader) Close() error {
	return dr.file.Close()
}
