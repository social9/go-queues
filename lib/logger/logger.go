package logger

import (
	"log"
	"os"

	"github.com/social9/go-queues/lib"
)

const (
	FATAL = iota
	ERROR
	WARN
	INFO
	DEBUG
)

// Config represents a logger configuration
//
// The defined levels are FATAL, ERROR, WARN, INFO, DEBUG  with values 0-4 in the same order. Set any other value to turn off logging.
//
// Level represents the maximum log level. For example, if it is set to 2, all logs below level 3 will be printed i.e. FATAL, ERROR, and WARN
//
// ** Please Note: The FATAL level causes the program to panic **
type Config struct {
	Level int // 0-5 : debug, info, warn, error, fatal
	Name  string
}

// Logger methods to call on a logger instance
type Logger interface {
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	Fatal(...interface{})
	SetLevel(level int)
	Child(opts Config) Logger

	print(caller int, msgs ...interface{})
}

// NewLogger retuns a logger instance
func NewLogger(opts Config) Logger {
	return &Config{opts.Level, opts.Name}
}

func (l *Config) print(caller int, msg ...interface{}) {
	name := l.Name

	var levelName string
	switch caller {
	case FATAL:
		levelName = "[FATAL]"
	case ERROR:
		levelName = "[ERROR]"
	case WARN:
		levelName = "[WARN]"
	case INFO:
		levelName = "[INFO]"
	case DEBUG:
		levelName = "[DEBUG]"
	default:
		return
	}

	var prefix []interface{}
	prefix = append(prefix, levelName, name)

	msg = append(prefix, msg...)
	log.Println(msg...)
}

// Debug log debug statements
func (l *Config) Debug(msg ...interface{}) {
	if l.Level == 0 {
		l.print(DEBUG, msg...)
	}
}

// Info log an info statement
func (l *Config) Info(msg ...interface{}) {
	if l.Level <= 1 {
		l.print(INFO, msg...)
	}
}

// Warn log a warn statement
func (l *Config) Warn(msg ...interface{}) {
	if l.Level <= 2 {
		l.print(WARN, msg...)
	}
}

// Error log a error statement
func (l *Config) Error(msg ...interface{}) {
	if l.Level <= 3 {
		l.print(ERROR, msg...)
	}
}

// Fatal log a fatal statement causing program to panic
func (l *Config) Fatal(msg ...interface{}) {
	if l.Level <= 4 {
		l.print(FATAL, msg...)
	}
	os.Exit(1)
}

// SetLevel change the verbose level
func (l *Config) SetLevel(level int) {
	l.Level = level
}

// Child create a child logger
//
// Defaults to the values set for parent logger
func (l *Config) Child(opts Config) Logger {
	level := lib.NonEmpty(opts.Level, l.Level).(int)
	name := "(" + opts.Name + ")"

	if l.Name != "" {
		name = "<" + l.Name + ">" + name
	}

	childLogger := NewLogger(Config{level, name})
	return childLogger
}
