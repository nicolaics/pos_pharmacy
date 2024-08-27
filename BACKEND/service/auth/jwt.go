package auth

import (
	"context"
	"fmt"
	_ "log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/nicolaics/pos_pharmacy/config"
	"github.com/nicolaics/pos_pharmacy/types"
	"github.com/nicolaics/pos_pharmacy/utils"
)

type contextKey string

const CashierKey contextKey = "cashierID"

func CreateJWT(cashierId int) (*types.TokenDetails, error) {
	tokenDetails := new(types.TokenDetails)

	accessExp := time.Second * time.Duration(config.Envs.JWTAccessExpirationInSeconds)

	tokenDetails.AccessTokenExp = time.Now().Add(accessExp).Unix()

	tempUUID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	tokenDetails.AccessUUID = tempUUID.String()

	tokenDetails.RefreshTokenExp = time.Now().Add(time.Hour * 24 * 7).Unix()

	tempUUID, err = uuid.NewV7()
	if err != nil {
		return nil, err
	}
	tokenDetails.RefreshToken = tempUUID.String()

	//Creating Access Token
	accessSecret := []byte(config.Envs.JWTAccessSecret)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized": true,
		"accessUuid": tokenDetails.AccessUUID,
		"cashierId":  cashierId,
		// expired of the token
		"expiredAt": tokenDetails.AccessTokenExp,
	})
	tokenDetails.AccessToken, err = accessToken.SignedString(accessSecret)
	if err != nil {
		return nil, err
	}

	//Creating Refresh Token
	refreshSecret := []byte(config.Envs.JWTRefreshSecret)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"refreshUuid": tokenDetails.RefreshUUID,
		"cashierId":   cashierId,
		// expired of the token
		"expiredAt": tokenDetails.RefreshTokenExp,
	})
	tokenDetails.RefreshToken, err = refreshToken.SignedString(refreshSecret)
	if err != nil {
		return nil, err
	}

	return tokenDetails, nil
}

func ValidateToken(r *http.Request) error {
	token, err := verifyToken(r)
	if err != nil {
		return err
	}

	_, ok := token.Claims.(jwt.Claims)

	if !ok && !token.Valid {
		return err
	}

	return nil
}

func verifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := extractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(config.Envs.JWTAccessSecret), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")

	//normally Authorization the_token_xxx

	strArr := strings.Split(bearerToken, " ")

	if len(strArr) == 2 {
		return strArr[1]
	}

	return ""
}

// func WithJWTAuth(handlerFunc http.HandlerFunc, store types.CashierStore) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		// get the token from the user request
// 		tokenString := getTokenFromRequest(r)

// 		// valdiate the JWT token
// 		token, err := validateToken(tokenString)
// 		if err != nil {
// 			log.Printf("failed to validate token: %v", err)
// 			permissionDenied(w)
// 			return
// 		}

// 		if !token.Valid {
// 			log.Println("invalid token")
// 			permissionDenied(w)
// 			return
// 		}

// 		// if it is correct, fetch the userID from the db
// 		claims := token.Claims.(jwt.MapClaims)
// 		str := claims["cashierID"].(string)

// 		cashierID, _ := strconv.Atoi(str)
// 		user, err := store.GetCashierByID(cashierID)

// 		if err != nil {
// 			log.Printf("failed to get user by id: %v", err)
// 			permissionDenied(w)
// 			return
// 		}

// 		// set the context "userID" to the userID
// 		ctx := r.Context()
// 		ctx = context.WithValue(ctx, CashierKey, user.ID)
// 		r = r.WithContext(ctx)

// 		handlerFunc(w, r)
// 	}
// }

func validateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(config.Envs.JWTAccessSecret), nil
	})
}

func ExtractTokenFromRedis(r *http.Request) (*types.AccessDetails, error) {
	token, err := verifyToken(r)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["accessUuid"].(string)
		if !ok {
			return nil, err
		}

		cashierId, err := strconv.Atoi(fmt.Sprintf("%.f", claims["cashierId"]))
		if err != nil {
			return nil, err
		}

		return &types.AccessDetails{
			AccessUUID: accessUuid,
			CashierID:  cashierId,
		}, nil
	}

	return nil, err
}

// func verifyToken(tokenString string) error {
// 	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 	   return secretKey, nil
// 	})

// 	if err != nil {
// 	   return err
// 	}

// 	if !token.Valid {
// 	   return fmt.Errorf("invalid token")
// 	}

// 	return nil
//  }

func permissionDenied(w http.ResponseWriter) {
	utils.WriteError(w, http.StatusForbidden, fmt.Errorf("permission denied"))
}

func GetUserIDFromContext(ctx context.Context) int {
	userID, ok := ctx.Value(CashierKey).(int)

	if !ok {
		return -1
	}

	return userID
}
