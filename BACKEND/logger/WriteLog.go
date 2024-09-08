package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

/*
logType = ["delete", "modify"]
logDataType = ["user", "invoice", "prescription", etc]
*/
func WriteLog(logType string, logDataType string, userId int, dataId int, deletedData any) error {
	logFolder := fmt.Sprintf("static/log/%s/%s", logType, logDataType)
	if err := os.MkdirAll(logFolder, 0755); err != nil {
		return err
	}

	currentDate := time.Now().Format("060102-1504") // YYMMDD 형식

	fileName := fmt.Sprintf("%s/%s_%d_%d.log", logFolder, currentDate, userId, dataId)

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