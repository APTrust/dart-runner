package util

import (
	"log"
	"os"
	"path"
)

type LogLevel int

const (
	LevelTrace LogLevel = -1
	LevelDebug LogLevel = 0
	LevelInfo  LogLevel = 1
	LevelWarn  LogLevel = 2
	LevelError LogLevel = 3
	LevelFatal LogLevel = 4
)

// Logger performs basic logging functions. We use this rather than
// Wails' built-in log because the built-in log is attached to the
// Wails context, which is not initiated until you call Wails.run().
// That makes unit testing difficult because it requires you to
// run the full graphical app and then pass Wails' internal context
// around to functions that probably shouldn't have direct access to
// it. This solution is cleaner and more flexible.
type Logger struct {
	log   *log.Logger
	level LogLevel
}

func GetLogger(level LogLevel) *Logger {
	paths := NewPaths()
	logFile := path.Join(paths.LogDir, "dart.log")
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return &Logger{
		log:   log.New(f, "", log.LstdFlags),
		level: level,
	}
}
func (log *Logger) SetLogLevel(level LogLevel) {
	log.level = level
}

func (l *Logger) Trace(msg string, args ...interface{}) {
	if l.level <= LevelTrace {
		l.log.Printf("[TRACE] "+msg, args...)
	}
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.log.Printf("[DEBUG] "+msg, args...)
	}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.log.Printf("[INFO] "+msg, args...)
	}
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.log.Printf("[WARN] "+msg, args...)
	}
}

func (l *Logger) Error(msg string, args ...interface{}) {
	if l.level <= LevelError {
		l.log.Printf("[ERROR] "+msg, args...)
	}
}

func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.log.Fatalf(msg, args...)
}

func (l *Logger) Panic(arg interface{}) {
	l.log.Panic(arg)
}
