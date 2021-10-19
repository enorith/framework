package database

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Logger struct {
	logLevel      logger.LogLevel
	logger        *zap.Logger
	SlowThreshold time.Duration
}

func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	l.logLevel = level
	return l
}

func (l *Logger) Info(_ context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		l.logger.Info(fmt.Sprintf(msg, data...))
	}
}

func (l *Logger) Warn(_ context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		l.logger.Warn(fmt.Sprintf(msg, data...))
	}
}

func (l *Logger) Error(_ context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		l.logger.Error(fmt.Sprintf(msg, data...))
	}
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if err == gorm.ErrRecordNotFound {
		return
	}
	if l.logLevel > 0 {
		elapsed := time.Since(begin)
		sql, rows := fc()
		fields := []zap.Field{
			zap.String("sql", sql),
			zap.Int64("rows_affected", rows),
			zap.String("duration", elapsed.String()),
		}
		log := l.logger.WithOptions(zap.AddCallerSkip(3))

		switch {
		case err != nil && l.logLevel >= logger.Error:
			log.Error(err.Error(), fields...)
		case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.logLevel >= logger.Warn:
			log.Warn("slow sql warning", fields...)
		case l.logLevel >= logger.Info:
			log.Info("", fields...)
		}
	}
}
