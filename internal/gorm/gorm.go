// This file contains the custom Gorm logger
// Origin Source: https://github.com/go-gorm/gorm/blob/master/logger/logger.go
// Author: Jinzhu
// Changes: Modified functions related to logging
// Licensed under the MIT License
/*
The MIT License (MIT)

Copyright (c) 2013-NOW  Jinzhu <wosmvp@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package gorm

import (
	"context"
	"fmt"
	internalLogger "github.com/openinfradev/tks-api/pkg/log"
	"github.com/pkg/errors"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

func (l *customGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

func (l *customGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		internalLogger.Infof(ctx, l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *customGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		internalLogger.Warnf(ctx, l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *customGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		internalLogger.Errorf(ctx, l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *customGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			internalLogger.Errorf(ctx, l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			//l.Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			internalLogger.Errorf(ctx, l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			//l.Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			internalLogger.Warnf(ctx, l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			//l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			internalLogger.Warnf(ctx, l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			//l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		if rows == -1 {
			internalLogger.Infof(ctx, l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
			//l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			internalLogger.Infof(ctx, l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
			//l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	default:
		internalLogger.Info(ctx, "Trace")
	}
}

type customGormLogger struct {
	logger.Writer
	logger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
	YellowBold  = "\033[33;1m"
)

func NewGormLogger() logger.Interface {
	infoStr := Green + "%s\n" + Reset + Green + "[info] " + Reset
	warnStr := BlueBold + "%s\n" + Reset + Magenta + "[warn] " + Reset
	errStr := Magenta + "%s\n" + Reset + Red + "[error] " + Reset
	traceStr := Green + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
	traceWarnStr := Green + "%s " + Yellow + "%s\n" + Reset + RedBold + "[%.3fms] " + Yellow + "[rows:%v]" + Magenta + " %s" + Reset
	traceErrStr := RedBold + "%s " + MagentaBold + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"

	return &customGormLogger{
		Writer: customGormLogger{},
		Config: logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  false,
		},
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceErrStr:  traceErrStr,
		traceWarnStr: traceWarnStr,
	}
}
