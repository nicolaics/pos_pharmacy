package auth

import (
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)


func AuthMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("auth middleware ok!")

			// log.Println(r.Header)

			token, err := verifyToken(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)

			if ok && token.Valid {
				_, ok := claims["tokenUuid"].(string)
				if !ok {
					http.Error(w, "error token uuid", http.StatusForbidden)
					return
				}

				_, ok = claims["userId"]

				if !ok {
					http.Error(w, "error user id", http.StatusForbidden)
					return
				}

				next.ServeHTTP(w, r)
			}
		})
	}
}

func CorsMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("cors middleware ok!")

			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE, PATCH")

			// Handle preflight (OPTIONS) request by returning 200 OK with the necessary headers
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Call the next handler in the chain (e.g., your AuthMiddleware)
			next.ServeHTTP(w, r)
		})
	}
}
