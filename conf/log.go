package conf

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/sirupsen/logrus"
)

// Log represents all the parameters needed for logging
type Log struct {
	Level string
	File  string
	Type  string
}

// InitLogger will take the logging configuration and also adds
// a few default parameters
func InitLogger(l *Log) (*logrus.Entry, error) {

	user, err := user.Current()
	if err != nil {
		return nil, err
	}
	userName := user.Name
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	// use a file if you want
	if l.File != "" {
		f, errOpen := os.OpenFile(l.File, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
		if errOpen != nil {
			return nil, errOpen
		}
		logrus.SetOutput(f)
		fmt.Println("Using output log file: ", l.File)
	}

	level, err := logrus.ParseLevel(strings.ToUpper(l.Level))
	if err != nil {
		return nil, err
	}
	logrus.SetLevel(level)

	var logFields logrus.Fields
	if l.Level == "debug" {
		logFields = logrus.Fields{"hostname": hostname, "user": userName}
	}

	if l.Type == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			DisableTimestamp: false,
		})

	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:    true,
			DisableTimestamp: false,
		})

	}

	return logrus.StandardLogger().WithFields(logFields), nil
}
