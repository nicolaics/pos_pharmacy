package utils

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Log     string      `json:"log,omitempty"`
	LogData string      `json:"logData,omitempty"`
	Error   string      `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

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

func (r *Response) WriteError(w http.ResponseWriter) {
	r.Success = false
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Code)
	json.NewEncoder(w).Encode(r)
	// WriteJSON(w, status, map[string]string{"error": err.Error()})
}

func (r *Response) WriteSuccess(w http.ResponseWriter) {
	r.Success = true
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Code)
	json.NewEncoder(w).Encode(r)
	// WriteJSON(w, status, data)
}
