package log

import (
	"context"
	"io"
	"os"
	"path"
	"runtime"
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
	logger.SetFormatter(&CustomFormatter{&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableQuote:    true,
	}})
	//logger.SetReportCaller(true)

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

type CustomFormatter struct {
	logrus.Formatter
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	//_, file := path.Split(entry.Caller.File)
	//entry.Data["file"] = file
	return f.Formatter.Format(entry)
}

// [TODO] more pretty
func Info(v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file": path.Base(file),
			"line": line,
		}).Info(v...)
	} else {
		logger.Info(v...)
	}
}
func Infof(format string, v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file": path.Base(file),
			"line": line,
		}).Infof(format, v...)
	} else {
		logger.Infof(format, v...)
	}
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
	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file": file,
			"line": line,
		}).Warn(v...)
	} else {
		logger.Warn(v...)
	}
}
func Warnf(format string, v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file": file,
			"line": line,
		}).Warnf(format, v...)
	} else {
		logger.Warnf(format, v...)
	}
}
func WarnWithContext(ctx context.Context, v ...interface{}) {
	reqID := ctx.Value(internal.ContextKeyRequestID)
	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file":                               file,
			"line":                               line,
			string(internal.ContextKeyRequestID): reqID,
		}).Warn(v...)
	} else {
		logger.WithField(string(internal.ContextKeyRequestID), reqID).Warn(v...)
	}
}
func WarnfWithContext(ctx context.Context, format string, v ...interface{}) {
	reqID := ctx.Value(internal.ContextKeyRequestID)
	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file":                               file,
			"line":                               line,
			string(internal.ContextKeyRequestID): reqID,
		}).Warnf(format, v...)
	} else {
		logger.WithField(string(internal.ContextKeyRequestID), reqID).Warnf(format, v...)
	}
}

func Debug(v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file": file,
			"line": line,
		}).Debug(v...)
	} else {
		logger.Debug(v...)
	}
}
func Debugf(format string, v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file": file,
			"line": line,
		}).Debugf(format, v...)
	} else {
		logger.Debugf(format, v...)
	}
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
	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file": file,
			"line": line,
		}).Error(v...)
	} else {
		logger.Error(v...)
	}
}
func Errorf(format string, v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file": file,
			"line": line,
		}).Errorf(format, v...)
	} else {
		logger.Errorf(format, v...)
	}
}
func ErrorWithContext(ctx context.Context, v ...interface{}) {
	reqID := ctx.Value(internal.ContextKeyRequestID)
	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file":                               file,
			"line":                               line,
			string(internal.ContextKeyRequestID): reqID,
		}).Error(v...)
	} else {
		logger.WithField(string(internal.ContextKeyRequestID), reqID).Error(v...)
	}
}
func ErrorfWithContext(ctx context.Context, format string, v ...interface{}) {
	reqID := ctx.Value(internal.ContextKeyRequestID)
	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file":                               file,
			"line":                               line,
			string(internal.ContextKeyRequestID): reqID,
		}).Errorf(format, v...)
	} else {
		logger.WithField(string(internal.ContextKeyRequestID), reqID).Errorf(format, v...)
	}
}

func Fatal(v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file": file,
			"line": line,
		}).Fatal(v...)
	} else {
		logger.Fatal(v...)
	}
}
func Fatalf(format string, v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		logger.WithFields(logrus.Fields{
			"file": file,
			"line": line,
		}).Fatalf(format, v...)
	} else {
		logger.Fatalf(format, v...)
	}
}

func Disable() {
	logger.Out = io.Discard
}
