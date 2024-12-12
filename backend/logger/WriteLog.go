package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nicolaics/pharmacon/utils"
)

/*
logType = ["delete", "modify"]
logDataType = ["user", "invoice", "prescription", etc]
*/
func WriteServerLog(logType string, logDataType string, userName string, dataId int, deletedData any) error {
	logFolder := fmt.Sprintf("static/log/%s/%s", logType, logDataType)
	if err := os.MkdirAll(logFolder, 0755); err != nil {
		return err
	}

	currentDate := time.Now().Format("060102-T-150405") // YYMMDD 형식

	fileName := fmt.Sprintf("%s/%s_%s_%d.log", logFolder, currentDate, userName, dataId)

	// 로그 파일 열기
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// JSON으로 변환
	jsonData, err := json.Marshal(deletedData)
	if err != nil {
		return err
	}

	// 파일에 로그 기록
	_, err = file.WriteString(fmt.Sprintf("%s\n", jsonData))
	if err != nil {
		return err
	}

	return nil
}

func WriteServerErrorLog(routes string, userId int, data any, errorMsg error) (string, error) {
	log.Println(errorMsg)

	logFolder := fmt.Sprintf("static/log/error/%s", time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(logFolder, 0755); err != nil {
		return "", err
	}

	currentDate := time.Now().Format("060102-150405") // YYMMDD-HHmmss
	fileName := fmt.Sprintf("%s-%s", currentDate, utils.GenerateRandomCodeAlphanumeric(6))
	filePath := fmt.Sprintf("%s/%s.log", logFolder, fileName)

	// 로그 파일 열기
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var jsonData []byte

	if data != nil {
		jsonData, _ = json.Marshal(data)
	}

	// store the data into the file
	msg := fmt.Sprintf("[Error %s]\n", time.Now().Format("2006/01/02 15:04:05"))
	msg += fmt.Sprintf("Routes: %s\n", routes)
	msg += fmt.Sprintf("User: %d\n", userId)
	msg += fmt.Sprintf("Additional Data: %s\n\n", string(jsonData))
	msg += fmt.Sprintf("%v\n", errorMsg)
	msg += "\n-----------------------------------------------------------------------------------------------\n\n"

	_, err = file.WriteString(msg)
	if err != nil {
		log.Printf("error write string: %v", err)
		return "", err
	}

	return fileName, nil
}
