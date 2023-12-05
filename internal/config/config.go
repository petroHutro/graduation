package config

import (
	"fmt"
	"time"
)

func NewFlags() Flags {
	return Flags{

		NetAddress: NetAddress{
			Host: "localhost",
			Port: 8080,
		},

		Logger: Logger{
			LoggerFilePath:  "file.log",
			LoggerFileFlag:  false,
			LoggerMultiFlag: false,
		},

		Storage: Storage{
			DatabaseDSN: "host=localhost user=url password=1234 dbname=url sslmode=disable",
		},

		Token: Token{
			TokenSecretKey: "",
			TokenTime: TokenTime{
				Time:     3,
				TokenEXP: time.Hour * 3,
			},
		},

		TicketKey: TicketKey{
			TicketSecretKey: "",
		},
	}
}

func LoadServerConfigure() (*Flags, error) {
	flags := parseFlags()
	parseENV(flags)
	if err := parseFile(flags); err != nil {
		return nil, fmt.Errorf("cannot pars file: %w", err)
	}
	return flags, nil
}
