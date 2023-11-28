package config

import (
	"flag"
)

func parseFlags() *Flags {
	flags := NewFlags()

	flag.Var(&flags.NetAddress, "a", "address and port to run server")

	flag.Var(&flags.TokenTime, "t", "user token lifetimer")

	flag.StringVar(&flags.TokenSecretKey, "k", "supersecretkey", "secret key for encoding the token")

	flag.StringVar(&flags.DatabaseDSN, "d", "host=localhost user=url password=1234 dbname=dbbot sslmode=disable", "DatabaseDSN")

	flag.BoolVar(&flags.Logger.LoggerFileFlag, "l", false, "Logger only file")
	flag.BoolVar(&flags.Logger.LoggerMultiFlag, "L", false, "Logger Multi")

	flag.Parse()
	return &flags
}
