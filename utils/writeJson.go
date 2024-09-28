package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, PATCH, GET, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With,Content-Type,Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length,Content-Range")
	w.WriteHeader(status)

	log.Println("JSON")
	log.Println(w.Header())

	return json.NewEncoder(w).Encode(v)
}

func WriteJSONForOptions(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, PATCH, GET, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With,Content-Type,Authorization")
	w.Header().Set("Access-Control-Max-Age", "1728000")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)

	log.Println("JSON Options")
	log.Println(w)

	return json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, map[string]string{"error": err.Error()})
}
