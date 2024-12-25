package env

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func GetString(key, fallback string) string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return fallback
	}
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func GetInt(key string, fallback int) int {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return fallback
	}
	val := os.Getenv(key)
	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		log.Fatal(err)
		return fallback
	}
	return valAsInt
}
