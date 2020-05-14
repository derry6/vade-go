package log

import (
    stdlog "log"
    "os"
)

var (
    _logger Logger = newDefault()
)

type Logger interface {
    Debugf(format string, args ...interface{})
    Infof(format string, args ...interface{})
    Warnf(format string, args ...interface{})
    Errorf(format string, args ...interface{})
    Fatalf(format string, args ...interface{})
}

type defaultLogger struct {
    std *stdlog.Logger
}

func newDefault() *defaultLogger {
    return &defaultLogger{
        stdlog.New(os.Stderr, "", stdlog.LstdFlags),
    }
}

func Get() Logger {
    return _logger
}

func Use(logger Logger) {
    if logger != nil {
        _logger = logger
    }
}

func (l *defaultLogger) Debugf(foramt string, args ...interface{}) {
    l.std.Printf(foramt, args...)
}
func (l *defaultLogger) Infof(foramt string, args ...interface{}) {
    l.std.Printf(foramt, args...)
}
func (l *defaultLogger) Warnf(foramt string, args ...interface{}) {
    l.std.Printf(foramt, args...)
}
func (l *defaultLogger) Errorf(foramt string, args ...interface{}) {
    l.std.Printf(foramt, args...)
}
func (l *defaultLogger) Fatalf(foramt string, args ...interface{}) {
    l.std.Fatalf(foramt, args...)
}
