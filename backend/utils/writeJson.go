package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, PATCH, GET, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With,Content-Type,Authorization, Response-Type")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length,Content-Range")
	w.WriteHeader(status)

	// log.Println(w.Header())

	return json.NewEncoder(w).Encode(v)
}

func WriteJSONForOptions(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, PATCH, GET, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With,Content-Type,Authorization,Response-Type")
	w.Header().Set("Access-Control-Max-Age", "1728000")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)

	// log.Println("JSON Options")
	// log.Println(w)

	return json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, err error, logFiles []string) {
	logMsg := ""

	if logFiles != nil {
		logMsg = "please contact administrator!\n"
		for _, logFile := range(logFiles) {
			logMsg += fmt.Sprintf("log: %s\n", logFile)
		}
	}

	response := map[string]interface{}{
		"response": err.Error(),
		"log":      logMsg,
	}
	WriteJSON(w, status, response)
}

func WriteSuccess(w http.ResponseWriter, status int, data any, logFiles []string) {
	logMsg := ""

	if logFiles != nil {
		logMsg = "please contact administrator!\n"
		for _, logFile := range(logFiles) {
			logMsg += fmt.Sprintf("log: %s\n", logFile)
		}
	}

	response := map[string]interface{}{
		"response": data,
		"log":      logMsg,
	}
	WriteJSON(w, status, response)
}

func WriteLog(w http.ResponseWriter, status int, data any, logFile string) {
	response := map[string]interface{}{
		"status":   "log",
		"response": data,
		"log":      logFile,
	}
	WriteJSON(w, status, response)
}
