package logger

import (
	"os"
	"strings"
	"time"

	"github.com/mzulfanw/boilerplate-go-fiber/internal/config"
	"github.com/sirupsen/logrus"
)

func Init(cfg config.Config) *logrus.Logger {
	log := logrus.StandardLogger()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})

	level, err := logrus.ParseLevel(strings.ToLower(cfg.LogLevel))
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	return log
}
