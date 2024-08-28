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

	err = token.Claims.Valid()

	if err != nil && !token.Valid {
		return err
	}

	return nil
}

func verifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString, err := extractToken(r)
	if err != nil {
		return nil, fmt.Errorf("unable to verify token")
	}

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

func extractToken(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")

	//normally: Authorization the_token_xxx
	strArr := strings.Split(token, " ")

	if len(strArr) == 2 {
		return strArr[1], nil
	}

	return "", fmt.Errorf("invalid token")
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
