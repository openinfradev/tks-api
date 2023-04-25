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

// Infof logs with format in InfoLevel.
func Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

// Warn logs in WarnLevel.
func Warn(v ...interface{}) {
	logger.Warn(v...)
}

// Warnf logs with format in WarnLevel.
func Warnf(format string, v ...interface{}) {
	logger.Warnf(format, v...)
}

// Debug logs in DebugLevel.
func Debug(v ...interface{}) {
	logger.Debug(v...)
}

// Debugf logs with format in DebugLevel.
func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

// Error logs in ErrorLevel.
func Error(v ...interface{}) {
	logger.Error(v...)
}

// Errorf logs with format in ErrorLevel.
func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

// Fatal logs in FatalLevel
func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

// Fatalf logs with format in FatalLevel
func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

func Disable() {
	logger.Out = ioutil.Discard
}
