package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
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

var colors = map[logrus.Level]string{
	logrus.DebugLevel: "\033[36m", // Cyan
	logrus.InfoLevel:  "\033[32m", // Green
	logrus.WarnLevel:  "\033[33m", // Yellow
	logrus.ErrorLevel: "\033[31m", // Red
	logrus.FatalLevel: "\033[31m", // Red
	logrus.PanicLevel: "\033[31m", // Red
}

func getColor(level logrus.Level) string {
	return colors[level]
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	levelColor := getColor(entry.Level)
	resetColor := "\033[0m"
	levelText := strings.ToUpper(entry.Level.String())

	requestIDColorStart := "\033[34m" // 파란색 시작
	requestIDColorEnd := "\033[0m"    // 색상 리셋

	requestID := entry.Data[string(internal.ContextKeyRequestID)]
	if requestID == nil {
		requestID = "Unknown"
	} else {
		requestID = fmt.Sprintf("%s%v%s", requestIDColorStart, requestID, requestIDColorEnd)
	}
	file := entry.Data["file"]
	if file == nil {
		file = "-"
	}
	var logMessage string

	if file == "-" {
		logMessage = fmt.Sprintf("%s%-7s%s %s %sREQUEST_ID=%v%s msg=%s\n",
			levelColor, levelText, resetColor,
			entry.Time.Format("2006-01-02 15:04:05"),
			requestIDColorStart, requestID, requestIDColorEnd,
			entry.Message,
		)
	} else {
		logMessage = fmt.Sprintf("%s%-7s%s %s %sREQUEST_ID=%v%s msg=%s file= %v\n",
			levelColor, levelText, resetColor,
			entry.Time.Format("2006-01-02 15:04:05"),
			requestIDColorStart, requestID, requestIDColorEnd,
			entry.Message,
			file,
		)
	}

	return []byte(logMessage), nil
}

// [TODO] more pretty
func Info(v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		logger.WithFields(logrus.Fields{
			"file": relativePath + ":" + strconv.Itoa(line),
		}).Info(v...)
	} else {
		logger.Info(v...)
	}
}
func Infof(format string, v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		logger.WithFields(logrus.Fields{
			"file": relativePath + ":" + strconv.Itoa(line),
		}).Infof(format, v...)
	} else {
		logger.Infof(format, v...)
	}
}
func InfoWithContext(ctx context.Context, v ...interface{}) {
	logger.WithField(string(internal.ContextKeyRequestID), ctx.Value(internal.ContextKeyRequestID)).Info(v...)
}
func InfofWithContext(ctx context.Context, format string, v ...interface{}) {
	logger.WithField(string(internal.ContextKeyRequestID), ctx.Value(internal.ContextKeyRequestID)).Infof(format, v...)
}

func Warn(v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		logger.WithFields(logrus.Fields{
			"file": relativePath + ":" + strconv.Itoa(line),
		}).Warn(v...)
	} else {
		logger.Warn(v...)
	}
}
func Warnf(format string, v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		logger.WithFields(logrus.Fields{
			"file": relativePath + ":" + strconv.Itoa(line),
		}).Warnf(format, v...)
	} else {
		logger.Warnf(format, v...)
	}
}
func WarnWithContext(ctx context.Context, v ...interface{}) {
	logger.WithField(string(internal.ContextKeyRequestID), ctx.Value(internal.ContextKeyRequestID)).Warn(v...)
}
func WarnfWithContext(ctx context.Context, format string, v ...interface{}) {
	logger.WithField(string(internal.ContextKeyRequestID), ctx.Value(internal.ContextKeyRequestID)).Warnf(format, v...)
}

func Debug(v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		logger.WithFields(logrus.Fields{
			"file": relativePath + ":" + strconv.Itoa(line),
		}).Debug(v...)
	} else {
		logger.Debug(v...)
	}
}
func Debugf(format string, v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		logger.WithFields(logrus.Fields{
			"file": relativePath + ":" + strconv.Itoa(line),
		}).Debugf(format, v...)
	} else {
		logger.Debugf(format, v...)
	}
}
func DebugWithContext(ctx context.Context, v ...interface{}) {
	logger.WithField(string(internal.ContextKeyRequestID), ctx.Value(internal.ContextKeyRequestID)).Debug(v...)
}
func DebugfWithContext(ctx context.Context, format string, v ...interface{}) {
	logger.WithField(string(internal.ContextKeyRequestID), ctx.Value(internal.ContextKeyRequestID)).Debugf(format, v...)
}

func Error(v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		logger.WithFields(logrus.Fields{
			"file": relativePath + ":" + strconv.Itoa(line),
		}).Error(v...)
	} else {
		logger.Error(v...)
	}
}
func Errorf(format string, v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		logger.WithFields(logrus.Fields{
			"file": relativePath + ":" + strconv.Itoa(line),
		}).Errorf(format, v...)
	} else {
		logger.Errorf(format, v...)
	}
}
func ErrorWithContext(ctx context.Context, v ...interface{}) {
	logger.WithField(string(internal.ContextKeyRequestID), ctx.Value(internal.ContextKeyRequestID)).Error(v...)
}
func ErrorfWithContext(ctx context.Context, format string, v ...interface{}) {
	logger.WithField(string(internal.ContextKeyRequestID), ctx.Value(internal.ContextKeyRequestID)).Errorf(format, v...)
}

func Fatal(v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		logger.WithFields(logrus.Fields{
			"file": relativePath + ":" + strconv.Itoa(line),
		}).Fatal(v...)
	} else {
		logger.Fatal(v...)
	}
}
func Fatalf(format string, v ...interface{}) {
	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		logger.WithFields(logrus.Fields{
			"file": relativePath + ":" + strconv.Itoa(line),
		}).Fatalf(format, v...)
	} else {
		logger.Fatalf(format, v...)
	}
}

func Disable() {
	logger.Out = io.Discard
}

func getRelativeFilePath(absolutePath string) string {
	wd, err := os.Getwd()
	if err != nil {
		return absolutePath
	}

	relativePath := strings.TrimPrefix(absolutePath, wd)

	if strings.HasPrefix(relativePath, "/") {
		relativePath = relativePath[1:]
	}

	return relativePath
}
