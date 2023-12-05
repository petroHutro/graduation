package config

import (
	"os"
)

func parseENV(flags *Flags) {
	if serverAddress := os.Getenv("SERVER_ADDRESS"); serverAddress != "" {
		flags.NetAddress.Set(serverAddress)
	}
	if time := os.Getenv("TOKEN_EXP"); time != "" {
		flags.TokenTime.Set(time)
	}
	if key := os.Getenv("SECRET_KEY"); key != "" {
		flags.TokenSecretKey = key
	}
	if fileLoggerPath := os.Getenv("LOGGER_FILE"); fileLoggerPath != "" {
		flags.LoggerFilePath = fileLoggerPath
	}
	if databaseDSN := os.Getenv("DATABASE_DSN"); databaseDSN != "" {
		flags.DatabaseDSN = databaseDSN
	}
	if tiketSecretKey := os.Getenv("SECRET_KEY_TICKET"); tiketSecretKey != "" {
		flags.TicketSecretKey = tiketSecretKey
	}
}
