package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) {
	tests := []struct {
		name                 string
		envServerAddr        string
		AccrualSystemAddress string
		expectedAddr         string
		expectedAcc          string
	}{
		{
			name:                 "Default values",
			envServerAddr:        "",
			AccrualSystemAddress: "",
			expectedAddr:         "localhost:8080",
			expectedAcc:          "127.0.0.1:8081",
		},
		{
			name:                 "Environment variables",
			envServerAddr:        "localhost:8082",
			AccrualSystemAddress: "localhost:8083",
			expectedAddr:         "localhost:8082",
			expectedAcc:          "localhost:8083",
		},
		{
			name:                 "Flags",
			envServerAddr:        "",
			AccrualSystemAddress: "",
			expectedAddr:         "localhost:8080",
			expectedAcc:          "127.0.0.1:8081",
		},
		{
			name:                 "Environment variables override flags",
			envServerAddr:        "localhost:8083",
			AccrualSystemAddress: "localhost:8084",
			expectedAddr:         "localhost:8083",
			expectedAcc:          "localhost:8084",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
			os.Args = []string{"cmd"}

			os.Setenv("RUN_ADDRESS", test.envServerAddr)
			os.Setenv("ACCRUAL_SYSTEM_ADDRESS", test.AccrualSystemAddress)

			cfg := NewConfig()

			assert.Equal(t, test.expectedAddr, cfg.RunAddress, "Неверный адрес сервера")
			assert.Equal(t, test.expectedAcc, cfg.AccrualSystemAddress, "Неверный адрес системы расчёта начислений")

			os.Unsetenv("RUN_ADDRESS")
			os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")
		})
	}
}
