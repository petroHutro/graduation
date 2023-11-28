package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func parsSMTP(flags *Flags) error {
	viper.SetConfigName("smtp-config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("cannot Read In Config smtp: %w", err)
	}

	// viper.SetDefault("smtp.smtpPort", 587)

	err = viper.Unmarshal(&flags.SMTP)
	if err != nil {
		return fmt.Errorf("cannot Unmarshal SMTP: %w", err)
	}

	viper.Reset()

	return nil
}

func parsObjectStorage(flags *Flags) error {
	viper.SetConfigName("objectstorage-config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("cannot Read In Config ObjectStorage: %w", err)
	}

	err = viper.Unmarshal(&flags.ObjectStorage)
	if err != nil {
		return fmt.Errorf("cannot Unmarshal ObjectStorage: %w", err)
	}

	viper.Reset()

	return nil
}

func parseFile(flags *Flags) error {
	if err := parsSMTP(flags); err != nil {
		return fmt.Errorf("cannot pars SMTP: %w", err)
	}

	if err := parsObjectStorage(flags); err != nil {
		return fmt.Errorf("cannot pars ObjectStorage: %w", err)
	}

	return nil
}
