package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Log *logrus.Logger = logrus.New()

func Initialize(level string) error {
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}

	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetOutput(os.Stdout)
	Log.SetLevel(logLevel)
	return nil
}
