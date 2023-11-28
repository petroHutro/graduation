package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type URLBase struct {
	BaseURL string
}

type Logger struct {
	LoggerFilePath  string
	LoggerFileFlag  bool
	LoggerMultiFlag bool
}

type ObjectStorage struct {
	StorageAccessKey  string `json:"storageAccessKey"`
	StorageSecretKey  string `json:"storageSecretKey"`
	StorageEndpoint   string `json:"storageEndpoint"`
	StorageBucketName string `json:"storageBucketName"`
	StorageUseSSL     bool   `json:"storageUseSSL"`
}

type Storage struct {
	DatabaseDSN string
	ObjectStorage
}

type NetAddress struct {
	Host string
	Port int
}

type TokenTime struct {
	TokenEXP time.Duration
	Time     int
}

type Token struct {
	TokenTime
	TokenSecretKey string
}

type SMTP struct {
	SmtpServer   string `json:"smtpServer"`
	SmtpUsername string `json:"smtpUsername"`
	SmtpPassword string `json:"smtpPassword"`
	From         string `json:"from"`
	SmtpPort     int    `json:"smtpPort"`
}

type Flags struct {
	NetAddress
	Logger
	Storage
	Token
	SMTP
}

func (a NetAddress) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *NetAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return fmt.Errorf("cannot atoi port: %w", err)
	}
	a.Host = hp[0]
	a.Port = port
	return nil
}

func (a TokenTime) String() string {
	return strconv.Itoa(a.Time)
}

func (a *TokenTime) Set(s string) error {
	hour, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("cannot atoi time: %w", err)
	}
	a.Time = hour
	a.TokenEXP = time.Hour * time.Duration(hour)
	return nil
}
