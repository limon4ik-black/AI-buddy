// Пакет logger — инициализация логгера из окружения.
package logger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/config"
)

// filePrefix + время запуска = имя файла: note-service-2026-11-12_18-05-33.log
const filePrefix = "note-service-"

// layout имени файла: без двоеточий, чтобы сортировался по алфавиту и не ломался на ФС.
const layout = "2006-01-02_15-04-05"

type Config struct {
	Level string
	Dir   string
}

// FromEnv читает LOG_LEVEL (info по умолчанию) и LOG_DIR (out/logs).
func FromEnv() Config {
	return Config{Level: config.Env("LOG_LEVEL", "info"), Dir: config.Env("LOG_DIR", "out/logs")}
}

// New собирает логгер: консоль плюс отдельный файл на каждый запуск.
// Если каталог недоступен, логгер всё равно рабочий — просто пишет только в консоль.
func New(cfg Config) (*zap.Logger, func()) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	enc := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "caller",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	encoder := func() zapcore.Encoder { return zapcore.NewConsoleEncoder(enc) }
	console := zapcore.NewCore(encoder(), zapcore.AddSync(os.Stdout), level)
	log := zap.New(console, zap.AddCaller())

	file, fileErr := open(cfg.Dir)
	if fileErr != nil {
		log.Warn("logging to file disabled", zap.String("dir", cfg.Dir), zap.Error(fileErr))
		return log, func() {}
	}
	log = zap.New(zapcore.NewTee(console, zapcore.NewCore(encoder(), zapcore.AddSync(file), level)), zap.AddCaller())
	log.Info("logging to file", zap.String("path", file.Name()))
	return log, func() { _ = file.Close() }
}

func open(dir string) (*os.File, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	name := filePrefix + time.Now().Format(layout) + ".log"
	return os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}
