package log

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger
)

// Init initializes logrus.logger and set
func init() {
	logger = logrus.New()

	logger.Out = os.Stdout

	formatter := new(logrus.TextFormatter)
	formatter.FullTimestamp = true
	formatter.TimestampFormat = "2006-01-02 15:04:05"
	logger.SetFormatter(formatter)

	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch logLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "warning":
		logger.SetLevel(logrus.WarnLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}
}

// Info logs in InfoLevel.
func Info(v ...interface{}) {
	logger.Info(v...)
}

// Warn logs in WarnLevel.
func Warn(v ...interface{}) {
	logger.Warn(v...)
}

// Debug logs in DebugLevel.
func Debug(v ...interface{}) {
	logger.Debug(v...)
}

// Error logs in ErrorLevel.
func Error(v ...interface{}) {
	logger.Error(v...)
}

// Fatal logs in FatalLevel
func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

func Disable() {
	logger.Out = ioutil.Discard
}
