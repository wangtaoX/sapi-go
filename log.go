package sapi

import (
	"github.com/Sirupsen/logrus"
	"os"
)

var (
	log = logrus.New()
)

func Log() (l *logrus.Logger) {
	return log
}

func InitLog(file string) error {
	fd, err := os.Create(file)
	if err != nil {
		return err
	}

	log.Out = fd
	return nil
}
