package log

import (
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Name   string
	Level  zapcore.Level
	Stdout bool
	File   string
}

func Init(c Config) (sync func() error, err error) {
	l := New(c)
	_ = zap.ReplaceGlobals(l)
	return l.Sync, nil
}

func AppendGlobal(name string) {
	l := zap.L()
	l = l.Named(name)
	_ = zap.ReplaceGlobals(l)
}

func New(c Config) (l *zap.Logger) {
	core := zapcore.NewTee()
	levelHandler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= c.Level
	})
	if c.Stdout {
		stdout := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.Lock(zapcore.AddSync(os.Stdout)),
			levelHandler,
		)
		core = zapcore.NewTee(core, stdout)
	}
	if c.File != "" {
		filecore := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(&lumberjack.Logger{
				Filename:   c.File,
				MaxSize:    500, // megabytes
				MaxBackups: 3,
				MaxAge:     28, // days
			}),
			levelHandler,
		)
		core = zapcore.NewTee(core, filecore)
	}
	l = zap.New(core).Named(c.Name)
	return l
}
