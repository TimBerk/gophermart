package config

import (
	"flag"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	LogLevel             string `env:"LOGGING_LEVEL" default:"info"`
	KeyJWT               []byte `env:"KEY_JWT" default:"gophermart"`
	ExpireJWT            int    `env:"EXPIRE_JWT" default:"60"`
}

func NewConfig() *Config {
	cfg := &Config{}
	err := envconfig.Process("gm", cfg)
	if err != nil {
		logrus.WithField("error", err).Error("Failed to process envconfig")
	}

	envServerAddress := os.Getenv("RUN_ADDRESS")
	envDatabaseURI := os.Getenv("DATABASE_URI")
	envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	envLogLevel := os.Getenv("LOGGING_LEVEL")
	envKeyJWT := os.Getenv("KEY_JWT")
	envExpireJWT := os.Getenv("EXPIRE_JWT")

	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Database URI for PostgreSQL")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "http://127.0.0.1:8081", "Base URL for accrual")
	flag.StringVar(&cfg.LogLevel, "l", "info", "Logging level")
	flag.Parse()

	if envServerAddress != "" {
		cfg.RunAddress = envServerAddress
	}
	if envDatabaseURI != "" {
		cfg.DatabaseURI = envDatabaseURI
	}
	if envAccrualSystemAddress != "" {
		cfg.AccrualSystemAddress = envAccrualSystemAddress
	}
	if envLogLevel != "" {
		cfg.LogLevel = envLogLevel
	}
	if envKeyJWT != "" {
		cfg.KeyJWT = []byte(envKeyJWT)
	}
	if envExpireJWT != "" {
		cfg.ExpireJWT, _ = strconv.Atoi(envExpireJWT)
	}
	return cfg
}
