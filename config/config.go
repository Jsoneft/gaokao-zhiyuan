package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port               string
	GinMode            string
	ClickHouseHost     string
	ClickHousePort     int
	ClickHouseUser     string
	ClickHousePassword string
	ClickHouseDatabase string
}

func LoadConfig() *Config {
	port := getEnv("PORT", "8031")
	ginMode := getEnv("GIN_MODE", "release")

	clickhousePort, _ := strconv.Atoi(getEnv("CLICKHOUSE_PORT", "19000"))

	return &Config{
		Port:               port,
		GinMode:            ginMode,
		ClickHouseHost:     getEnv("CLICKHOUSE_HOST", "localhost"),
		ClickHousePort:     clickhousePort,
		ClickHouseUser:     getEnv("CLICKHOUSE_USERNAME", "default"),
		ClickHousePassword: getEnv("CLICKHOUSE_PASSWORD", ""),
		ClickHouseDatabase: getEnv("CLICKHOUSE_DATABASE", "gaokao"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
