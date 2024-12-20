package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	PublicHost                 string
	Port                       string
	DBUser                     string
	DBPassword                 string
	DBAddress                  string
	DBName                     string
	JWTExpirationInSeconds     int64
	JWTSecret                  string
	CompanyName                string
	Pharmacist                 string
	PharmacistLicenseNumber    string
	MainDoctor                 string
	MainDoctorLicenseNumber    string
	BusinessRegistrationNumber string
	CompanyAddress             string
	CompanyPhoneNumber         string
	CompanyWhatsAppNumber      string
	CompanyLogoURL             string
	CompanySlogan              string
}

var Envs = initConfig()

func initConfig() Config {
	godotenv.Load()

	return Config{
		PublicHost: getEnv("PUBLIC_HOST", "http://localhost"),
		Port:       getEnv("PORT", "19230"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBAddress: fmt.Sprintf("%s:%s", getEnv("DB_HOST", "127.0.0.1"),
			getEnv("DB_PORT", "3306")),
		DBName:                     getEnv("DB_NAME", "pos_test"),
		JWTExpirationInSeconds:     getEnvAsInt("JWT_EXP", (3600 * 12)), // for 12 hours
		JWTSecret:                  getEnv("JWT_SECRET", "access-secret"),
		CompanyName:                getEnv("COMPANY_NAME", "Apotek"),
		Pharmacist:                 getEnv("PHARMACIST", ""),
		PharmacistLicenseNumber:    getEnv("PHARMACIST_LICENSE_NUMBER", ""),
		MainDoctor:                 getEnv("MAIN_DOCTOR", ""),
		MainDoctorLicenseNumber:    getEnv("MAIN_DOCTOR_LICENSE_NUMBER", ""),
		BusinessRegistrationNumber: getEnv("BUSINESS_REGISTRATION_NUMBER", ""),
		CompanyAddress:             getEnv("COMPANY_ADDRESS", ""),
		CompanyPhoneNumber:         getEnv("COMPANY_PHONE_NUMBER", ""),
		CompanyWhatsAppNumber:      getEnv("COMPANY_WHATSAPP_NUMBER", ""),
		CompanyLogoURL:             getEnv("COMPANY_LOGO_URL", "static/assets/logo/Logo.png"),
		CompanySlogan:              getEnv("COMPANY_SLOGAN", ""),
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
