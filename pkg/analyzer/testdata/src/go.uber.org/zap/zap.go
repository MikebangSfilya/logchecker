package zap

type Logger struct{}

func NewProduction() (*Logger, error) { return &Logger{}, nil }

func (l *Logger) Sync() error                     { return nil }
func (l *Logger) Info(msg string, fields ...any)  {}
func (l *Logger) Error(msg string, fields ...any) {}
func (l *Logger) Warn(msg string, fields ...any)  {}
func (l *Logger) Debug(msg string, fields ...any) {}
