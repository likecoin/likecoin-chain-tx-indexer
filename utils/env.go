package utils

import (
	"os"
	"strconv"
)

func Env(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func EnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if number, err := strconv.Atoi(value); err == nil {
			return number
		}
	}
	return defaultValue
}
