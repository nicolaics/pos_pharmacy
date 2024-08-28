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

func ValidateAccessToken(r *http.Request) error {
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

func ValidateRefreshToken(r *http.Request) (string, int, error) {
	token, err := verifyToken(r)
	if err != nil {
		return "", -1, err
	}

	err = token.Claims.Valid()

	if err != nil && !token.Valid {
		return "", -1, err
	}

	claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims

	if ok && token.Valid {
		refreshUuid, ok := claims["refresh_uuid"].(string) //convert the interface to string
		if !ok {
			return "", -1, err
		}

		cashierId, err := strconv.Atoi(fmt.Sprintf("%.f", claims["cashierId"]))
		if err != nil {
			return "", -1, err
		}

		return refreshUuid, cashierId, nil
	} else {
		return "", -1, fmt.Errorf("refresh expired")
	}
}

func verifyToken(r *http.Request) (*jwt.Token, error) {
	tokenStringArr, err := extractToken(r)
	if err != nil {
		return nil, fmt.Errorf("unable to verify token")
	}

	token, err := jwt.Parse(tokenStringArr[1], func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		if tokenStringArr[0] == "Access" {
			return []byte(config.Envs.JWTAccessSecret), nil
		} else {
			return []byte(config.Envs.JWTRefreshSecret), nil
		}
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func extractToken(r *http.Request) ([]string, error) {
	token := r.Header.Get("Authorization")

	//normally: Authorization the_token_xxx
	strArr := strings.Split(token, " ")

	if len(strArr) == 2 {
		return strArr, nil
	}

	return nil, fmt.Errorf("invalid token")
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
