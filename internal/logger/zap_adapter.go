package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapAdapter struct {
	core  *zap.Logger
	level zap.AtomicLevel
}

func (z *zapAdapter) Debug(msg string, fields ...Field) {
	z.core.Debug(msg, toZapFields(fields)...)
}

func (z *zapAdapter) Info(msg string, fields ...Field) {
	z.core.Info(msg, toZapFields(fields)...)
}

func (z *zapAdapter) Warn(msg string, fields ...Field) {
	z.core.Warn(msg, toZapFields(fields)...)
}

func (z *zapAdapter) Error(msg string, fields ...Field) {
	z.core.Error(msg, toZapFields(fields)...)
}

func (z *zapAdapter) With(fields ...Field) Logger {
	return &zapAdapter{core: z.core.With(toZapFields(fields)...), level: z.level}
}

func (z *zapAdapter) Sync() error {
	return z.core.Sync()
}

func (z *zapAdapter) Level() Level {
	return fromZapLevel(z.level.Level())
}

func (z *zapAdapter) SetLevel(level Level) {
	z.level.SetLevel(toZapLevel(level))
}

// fromZapLevel maps Zap's level back to custom Level.
func fromZapLevel(l zapcore.Level) Level {
	switch l {
	case zap.DebugLevel:
		return LevelDebug
	case zap.InfoLevel:
		return LevelInfo
	case zap.WarnLevel:
		return LevelWarn
	case zap.ErrorLevel:
		return LevelError
	default:
		return LevelInfo
	}
}

// toZapLevel maps custom Level to Zap's level.
func toZapLevel(l Level) zapcore.Level {
	switch l {
	case LevelDebug:
		return zap.DebugLevel
	case LevelInfo:
		return zap.InfoLevel
	case LevelWarn:
		return zap.WarnLevel
	case LevelError:
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}

func NewZapAdapter() (Logger, error) {
	config := zap.NewProductionConfig()
	config.Encoding = "json"
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.Sampling = &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	}
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	z, err := config.Build(zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		return nil, err
	}
	return &zapAdapter{core: z, level: config.Level}, nil
}

func toZapFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))

	for i, f := range fields {
		switch v := f.Value.(type) {
		case string:
			zapFields[i] = zap.String(f.Key, v)
		case int:
			zapFields[i] = zap.Int(f.Key, v)
		case error:
			zapFields[i] = zap.Error(v)
		default:
			zapFields[i] = zap.Any(f.Key, v)
		}
	}
	return zapFields
}
