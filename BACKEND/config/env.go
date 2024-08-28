package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	PublicHost                    string
	Port                          string
	DBUser                        string
	DBPassword                    string
	DBAddress                     string
	DBName                        string
	JWTAccessExpirationInSeconds  int64
	JWTRefreshExpirationInSeconds int64
	JWTAccessSecret               string
	JWTRefreshSecret              string
	RedisDSN                      string
}

var Envs = initConfig()

func initConfig() Config {
	godotenv.Load()

	return Config{
		PublicHost: getEnv("PUBLIC_HOST", "http://localhost"),
		Port:       getEnv("PORT", "19230"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", "root1234!"),
		DBAddress: fmt.Sprintf("%s:%s", getEnv("DB_HOST", "127.0.0.1"),
			getEnv("DB_PORT", "3306")),
		DBName:                        getEnv("DB_NAME", "pos_pharmacy"),
		JWTAccessExpirationInSeconds:  getEnvAsInt("JWT_ACCESS_EXP", (3600 * 12)),  // for seven days
		JWTRefreshExpirationInSeconds: getEnvAsInt("JWT_REFRESH_EXP", (3600 * 24)), // for seven days
		JWTAccessSecret:                     getEnv("JWT_ACCESS_SECRET", "nsjuwpiiaAjM"),
		JWTRefreshSecret:                     getEnv("JWT_REFRESH_SECRET", "euNwhiwpmql"),
		RedisDSN:                      getEnv("REDIS_DSN", "localhost:6379"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)

		if err != nil {
			return int64(fallback)
		}

		return i
	}

	return fallback
}
