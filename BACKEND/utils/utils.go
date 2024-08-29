package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}

	return json.NewDecoder(r.Body).Decode(payload)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "POST, OPTIONS, PATCH, GET, DELETE")
	w.Header().Add("Access-Control-Allow-Headers", "X-Requested-With,Content-Type,Authorization")
	w.Header().Add("Access-Control-Expose-Headers", "Content-Length,Content-Range")
	w.WriteHeader(status)

	log.Println("JSON")
	log.Println(w.Header())

	return json.NewEncoder(w).Encode(v)
}

func WriteJSONForOptions(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "POST, OPTIONS, PATCH, GET, DELETE")
	w.Header().Add("Access-Control-Allow-Headers", "X-Requested-With,Content-Type,Authorization")
	w.Header().Add("Access-Control-Max-Age", "1728000")
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)

	log.Println("JSON Options")
	log.Println(w)

	return json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, map[string]string{"error": err.Error()})
}
