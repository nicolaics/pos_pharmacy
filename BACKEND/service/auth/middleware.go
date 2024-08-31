package auth

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// )

// SRC: https://gist.github.com/AxelRHD/2344cc1105afc06723b363f21486dec8

// func AuthenticationMiddleware(next http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		err := ValidateAccessToken(r)
// 		if err != nil {
// 			http.Error(w, "authentication error!", http.StatusForbidden)
// 			log.Println("authentication error!")

// 			return
// 		}

// 		next(w, r)
// 	}
// }

// // restricted secret route
// func AuthHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintln(w, "authorized")
// }
