package log

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/openinfradev/tks-api/internal"
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

// [TODO] more pretty
func Info(v ...interface{}) {
	logger.Info(v...)
}
func Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}
func InfoWithContext(ctx context.Context, v ...interface{}) {
	reqID := ctx.Value(internal.ContextKeyRequestID)
	logger.WithField(string(internal.ContextKeyRequestID), reqID).Info(v...)
}
func InfofWithContext(ctx context.Context, format string, v ...interface{}) {
	reqID := ctx.Value(internal.ContextKeyRequestID)
	logger.WithField(string(internal.ContextKeyRequestID), reqID).Infof(format, v...)
}

func Warn(v ...interface{}) {
	logger.Warn(v...)
}
func Warnf(format string, v ...interface{}) {
	logger.Warnf(format, v...)
}
func WarnWithContext(ctx context.Context, v ...interface{}) {
	reqID := ctx.Value(internal.ContextKeyRequestID)
	logger.WithField(string(internal.ContextKeyRequestID), reqID).Warn(v...)
}
func WarnfWithContext(ctx context.Context, format string, v ...interface{}) {
	reqID := ctx.Value(internal.ContextKeyRequestID)
	logger.WithField(string(internal.ContextKeyRequestID), reqID).Warnf(format, v...)
}

func Debug(v ...interface{}) {
	logger.Debug(v...)
}
func Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}
func DebugWithContext(ctx context.Context, v ...interface{}) {
	reqID := ctx.Value(internal.ContextKeyRequestID)
	logger.WithField(string(internal.ContextKeyRequestID), reqID).Debug(v...)
}
func DebugfWithContext(ctx context.Context, format string, v ...interface{}) {
	reqID := ctx.Value(internal.ContextKeyRequestID)
	logger.WithField(string(internal.ContextKeyRequestID), reqID).Debugf(format, v...)
}

func Error(v ...interface{}) {
	logger.Error(v...)
}
func Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}
func ErrorWithContext(ctx context.Context, v ...interface{}) {
	reqID := ctx.Value(internal.ContextKeyRequestID)
	logger.WithField(string(internal.ContextKeyRequestID), reqID).Error(v...)
}
func ErrorfWithContext(ctx context.Context, format string, v ...interface{}) {
	reqID := ctx.Value(internal.ContextKeyRequestID)
	logger.WithField(string(internal.ContextKeyRequestID), reqID).Errorf(format, v...)
}

func Fatal(v ...interface{}) {
	logger.Fatal(v...)
}
func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

func Disable() {
	logger.Out = io.Discard
}
