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
func Info(ctx context.Context, v ...interface{}) {
	fields := logrus.Fields{}

	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		fields["file"] = relativePath + ":" + strconv.Itoa(line)
	}

	if ctx != nil {
		fields[string(internal.ContextKeyRequestID)] = ctx.Value(internal.ContextKeyRequestID)
	}

	logger.WithFields(fields).Info(v...)
}
func Infof(ctx context.Context, format string, v ...interface{}) {
	fields := logrus.Fields{}

	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		fields["file"] = relativePath + ":" + strconv.Itoa(line)
	}

	if ctx != nil {
		fields[string(internal.ContextKeyRequestID)] = ctx.Value(internal.ContextKeyRequestID)
	}

	logger.WithFields(fields).Infof(format, v...)
}

func Warn(ctx context.Context, v ...interface{}) {
	fields := logrus.Fields{}

	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		fields["file"] = relativePath + ":" + strconv.Itoa(line)
	}

	if ctx != nil {
		fields[string(internal.ContextKeyRequestID)] = ctx.Value(internal.ContextKeyRequestID)
	}

	logger.WithFields(fields).Warn(v...)
}
func Warnf(ctx context.Context, format string, v ...interface{}) {
	fields := logrus.Fields{}

	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		fields["file"] = relativePath + ":" + strconv.Itoa(line)
	}

	if ctx != nil {
		fields[string(internal.ContextKeyRequestID)] = ctx.Value(internal.ContextKeyRequestID)
	}

	logger.WithFields(fields).Warnf(format, v...)
}

func Debug(ctx context.Context, v ...interface{}) {
	fields := logrus.Fields{}

	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		fields["file"] = relativePath + ":" + strconv.Itoa(line)
	}

	if ctx != nil {
		fields[string(internal.ContextKeyRequestID)] = ctx.Value(internal.ContextKeyRequestID)
	}

	logger.WithFields(fields).Debug(v...)
}
func Debugf(ctx context.Context, format string, v ...interface{}) {
	fields := logrus.Fields{}

	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		fields["file"] = relativePath + ":" + strconv.Itoa(line)
	}

	if ctx != nil {
		fields[string(internal.ContextKeyRequestID)] = ctx.Value(internal.ContextKeyRequestID)
	}

	logger.WithFields(fields).Debugf(format, v...)
}

func Error(ctx context.Context, v ...interface{}) {
	fields := logrus.Fields{}

	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		fields["file"] = relativePath + ":" + strconv.Itoa(line)
	}

	if ctx != nil {
		fields[string(internal.ContextKeyRequestID)] = ctx.Value(internal.ContextKeyRequestID)
	}

	logger.WithFields(fields).Error(v...)
}
func Errorf(ctx context.Context, format string, v ...interface{}) {
	fields := logrus.Fields{}

	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		fields["file"] = relativePath + ":" + strconv.Itoa(line)
	}

	if ctx != nil {
		fields[string(internal.ContextKeyRequestID)] = ctx.Value(internal.ContextKeyRequestID)
	}

	logger.WithFields(fields).Errorf(format, v...)
}

func Fatal(ctx context.Context, v ...interface{}) {
	fields := logrus.Fields{}

	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		fields["file"] = relativePath + ":" + strconv.Itoa(line)
	}

	if ctx != nil {
		fields[string(internal.ContextKeyRequestID)] = ctx.Value(internal.ContextKeyRequestID)
	}

	logger.WithFields(fields).Fatal(v...)
}
func Fatalf(ctx context.Context, format string, v ...interface{}) {
	fields := logrus.Fields{}

	if _, file, line, ok := runtime.Caller(1); ok {
		relativePath := getRelativeFilePath(file)
		fields["file"] = relativePath + ":" + strconv.Itoa(line)
	}

	if ctx != nil {
		fields[string(internal.ContextKeyRequestID)] = ctx.Value(internal.ContextKeyRequestID)
	}

	logger.WithFields(fields).Fatalf(format, v...)
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
