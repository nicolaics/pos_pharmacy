package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}

	return json.NewDecoder(r.Body).Decode(payload)
}

// GenerateRandomCode 는 랜덤한 6자리의 이메일 인증코드를 생성합니다.
func GenerateRandomCodeNumbers(length int) string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	const charset = "0123456789"

	result := make([]byte, length)

	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}

func GenerateRandomCodeAlphanumeric(length int) string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	result := make([]byte, length)

	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}

// dateStr must include the GMT
func ParseDate(dateStr string) (*time.Time, error) {
	dateFormat := "2006-01-02 -0700MST"
	date, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return nil, err
	}

	return &date, nil
}

func ParseStartDate(dateStr string) (*time.Time, error) {
	dateFormat := "2006-01-02 -0700MST"
	date, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return nil, err
	}

	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	return &date, nil
}

func ParseEndDate(dateStr string) (*time.Time, error) {
	dateFormat := "2006-01-02 -0700MST"
	date, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return nil, err
	}

	date = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())

	return &date, nil
}