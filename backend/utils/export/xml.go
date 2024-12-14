package export

import (
	"encoding/xml"
	"os"
)

func CreateXmlData(fileName string, data any) (string, error) {
	directory := "static/export/xml/"
	if err := os.MkdirAll(directory, 0744); err != nil {
		return "", err
	}

	filePath := directory + fileName

	xmlFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer xmlFile.Close()
	
	xmlFile.WriteString(xml.Header)

	// output, err := xml.MarshalIndent(data, "", "\t")
	// if err != nil {
	// 	return "", err
	// }

	encoder := xml.NewEncoder(xmlFile)
	encoder.Indent("", "\t")
	
	err = encoder.Encode(&data)
	if err != nil {
		return "", err
	}
	
	err = encoder.Close()
	if err != nil {
		return "", err
	}

	return filePath, nil
}