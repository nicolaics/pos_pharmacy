package export

import (
	"encoding/csv"
	"os"
)

func CreateCsvWriter(fileName string) (string, *csv.Writer, *os.File, error) {
	directory := "static/export/csv/"
	if err := os.MkdirAll(directory, 0744); err != nil {
		return "", nil, nil, err
	}

	filePath := directory + fileName

	f, err := os.Create(filePath)
	if err != nil {
		return "", nil, nil, err
	}

	writer := csv.NewWriter(f)

	return filePath, writer, f, nil
}

func WriteCsvHeader(writer *csv.Writer, header []string) error {
	err := writer.Write(header)
	if err != nil {
		return err
	}

	writer.Flush()
	if writer.Error() != nil {
		return writer.Error()
	}

	return nil
}

func WriteCsvData(writer *csv.Writer, record []string) error {
	err := writer.Write(record)
	if err != nil {
		return err
	}

	writer.Flush()
	if writer.Error() != nil {
		return writer.Error()
	}

	return nil
}
