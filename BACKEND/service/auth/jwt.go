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

// TODO: remove refresh token
func CreateJWT(userId int) (*types.TokenDetails, error) {
	tokenDetails := new(types.TokenDetails)

	accessExp := time.Second * time.Duration(config.Envs.JWTAccessExpirationInSeconds)

	tokenDetails.TokenExp = time.Now().Add(accessExp).Unix()

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
		"userId":     userId,
		// expired of the token
		"expiredAt": tokenDetails.TokenExp,
	})
	tokenDetails.Token, err = accessToken.SignedString(accessSecret)
	if err != nil {
		return nil, err
	}

	//Creating Refresh Token
	refreshSecret := []byte(config.Envs.JWTRefreshSecret)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"refreshUuid": tokenDetails.RefreshUUID,
		"userId":      userId,
		// expired of the token
		"expiredAt": tokenDetails.RefreshTokenExp,
	})
	tokenDetails.RefreshToken, err = refreshToken.SignedString(refreshSecret)
	if err != nil {
		return nil, err
	}

	return tokenDetails, nil
}

func ExtractRefreshTokenFromClient(r *http.Request) (*types.RefreshDetails, error) {
	token, err := verifyRefreshToken(r)
	if err != nil {
		return nil, err
	}

	err = token.Claims.Valid()

	if err != nil && !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims

	if ok && token.Valid {
		refreshUuid, ok := claims["refreshUuid"].(string) //convert the interface to string
		if !ok {
			return nil, err
		}

		userId, err := strconv.Atoi(fmt.Sprintf("%.f", claims["userId"]))
		if err != nil {
			return nil, err
		}

		return &types.RefreshDetails{
			RefreshUUID: refreshUuid,
			UserID:      userId,
		}, nil
	}

	return nil, fmt.Errorf("refresh expired")
}

func verifyRefreshToken(r *http.Request) (*jwt.Token, error) {
	tokenStr, err := extractToken(r)
	if err != nil {
		return nil, fmt.Errorf("unable to verify token")
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(config.Envs.JWTRefreshSecret), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func ExtractTokenFromClient(r *http.Request) (*types.AccessDetails, error) {
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

		userId, err := strconv.Atoi(fmt.Sprintf("%.f", claims["userId"]))
		if err != nil {
			return nil, err
		}

		return &types.AccessDetails{
			AccessUUID: accessUuid,
			UserID:     userId,
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

		return []byte(config.Envs.JWTAccessSecret), nil
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
