package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/nicolaics/pos_pharmacy/config"
	"github.com/nicolaics/pos_pharmacy/types"
)

type contextKey string

const UserKey contextKey = "userID"

func CreateJWT(userId int) (*types.TokenDetails, error) {
	tokenDetails := new(types.TokenDetails)

	tokenExp := time.Second * time.Duration(config.Envs.JWTExpirationInSeconds)

	tokenDetails.TokenExp = time.Now().Add(tokenExp).Unix()

	tempUUID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	tokenDetails.UUID = tempUUID.String()

	//Creating Access Token
	tokenSecret := []byte(config.Envs.JWTSecret)
	tokenToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized": true,
		"tokenUuid":  tokenDetails.UUID,
		"userId":     userId,
		// expired of the token
		"expiredAt": tokenDetails.TokenExp,
	})
	tokenDetails.Token, err = tokenToken.SignedString(tokenSecret)
	if err != nil {
		return nil, err
	}

	return tokenDetails, nil
}

func ExtractTokenFromClient(r *http.Request) (*types.AccessDetails, error) {
	token, err := verifyToken(r)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if ok && token.Valid {
		tokenUuid, ok := claims["tokenUuid"].(string)
		if !ok {
			return nil, err
		}

		userId, err := strconv.Atoi(fmt.Sprintf("%.f", claims["userId"]))
		if err != nil {
			return nil, err
		}

		return &types.AccessDetails{
			UUID:   tokenUuid,
			UserID: userId,
		}, nil
	}

	return nil, err
}

func verifyToken(r *http.Request) (*jwt.Token, error) {
	tokenStr, err := extractToken(r)
	if err != nil {
		return nil, fmt.Errorf("unable to verify token")
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(config.Envs.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func extractToken(r *http.Request) (string, error) {
	tokenString := r.Header.Get("Authorization")

	//normally: Authorization the_token_xxx
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	if tokenString != "" {
		return tokenString, nil
	}

	return "", fmt.Errorf("invalid token")
}
