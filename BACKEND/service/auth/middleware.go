package auth

// TODO: CHECK SESSION WARE IS WORKING OR NOT
import (
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var (
	Store = sessions.NewCookieStore([]byte("test"))
)

// func SessionMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// 세션 검사
// 		session, _ := Store.Get(r, "session-name")

// 		auth, ok := session.Values["authenticated"].(bool)
// 		if !ok || !auth {
// 			http.Error(w, "Forbidden", http.StatusForbidden)
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

// TODO: CROSSCHECK SESSION MIDDLEWARE
func SessionMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 세션 검사
			log.Println("session running")

			session, _ := Store.Get(r, "session-name")

			auth, ok := session.Values["authenticated"].(bool)
			if !ok || !auth {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}


func CheckSession(r *http.Request, needAdmin bool) (bool, error) {
	log.Println("check session ok")

	session, err := Store.Get(r, "session-name")
	if err != nil {
		return false, fmt.Errorf("session invalid")
	}

	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		return false, fmt.Errorf("auth failed")
	}

	if needAdmin {
		admin, ok := session.Values["admin"].(bool)
		if !ok || !admin {
			return false, fmt.Errorf("not admin")
		}
	}

	return true, nil
}

func AuthMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("auth ware ok!")

			token, err := verifyToken(r)
			if err != nil {
				http.Error(w, "error verify token", http.StatusForbidden)
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
				log.Println("userID: ", claims["userId"])

				if !ok {
					http.Error(w, "error user id", http.StatusForbidden)
					return
				}

				// _, ok = claims["admin"].(bool)
				// if !ok {
				// 	http.Error(w, "error admin", http.StatusForbidden)
				// 	return
				// }
				
				next.ServeHTTP(w, r)
			}
		})
	}
}

