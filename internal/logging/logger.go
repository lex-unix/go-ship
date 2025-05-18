package logging

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
)

type Level int

const (
	LevelDebug Level = -4
	LevelInfo  Level = 0
	LevelWarn  Level = 4
	LevelError Level = 8
)

var (
	debugColor = color.New(color.FgMagenta).SprintFunc()
	infoColor  = color.New(color.FgGreen).SprintFunc()
	warnColor  = color.New(color.FgYellow).SprintFunc()
	errorColor = color.New(color.FgRed).SprintFunc()
	blueColor  = color.New(color.FgBlue).SprintFunc()
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return ""
	}
}

func (l Level) ColorString() string {
	levelStr := l.String()
	switch l {
	case LevelDebug:
		return debugColor(levelStr)
	case LevelInfo:
		return infoColor(levelStr)
	case LevelWarn:
		return warnColor(levelStr)
	case LevelError:
		return errorColor(levelStr)
	default:
		return levelStr // return uncolored if unknown
	}
}

var defaultLogger atomic.Pointer[Logger]

func init() {
	defaultLogger.Store(New(os.Stdout, LevelInfo))
}

type Logger struct {
	out   io.Writer
	mu    sync.Mutex
	level Level
}

func Default() *Logger { return defaultLogger.Load() }

func SetDefault(l *Logger) { defaultLogger.Store(l) }

func New(out io.Writer, level Level) *Logger {
	return &Logger{
		out:   out,
		mu:    sync.Mutex{},
		level: level,
	}
}

func Debug(msg string) {
	Default().log(LevelDebug, msg)
}

func Info(msg string) {
	Default().log(LevelInfo, msg)
}

func Warn(msg string) {
	Default().log(LevelWarn, msg)
}

func Error(msg string) {
	Default().log(LevelError, msg)
}

func DebugHost(host, msg string) {
	Default().logWithHost(LevelDebug, host, msg)
}

func InfoHost(host, msg string) {
	Default().logWithHost(LevelInfo, host, msg)
}

func WarnHost(host, msg string) {
	Default().logWithHost(LevelWarn, host, msg)
}

func ErrorHost(host, msg string) {
	Default().logWithHost(LevelError, host, msg)
}

func Debugf(format string, args ...any) {
	Default().log(LevelDebug, format, args...)
}

func Infof(format string, args ...any) {
	Default().log(LevelInfo, format, args...)
}

func Warnf(format string, args ...any) {
	Default().log(LevelWarn, format, args...)
}

func Errorf(format string, args ...any) {
	Default().log(LevelError, format, args...)
}

func DebugHostf(host, format string, args ...any) {
	Default().logWithHost(LevelDebug, host, format, args...)
}

func InfoHostf(host, format string, args ...any) {
	Default().logWithHost(LevelInfo, host, format, args...)
}

func WarnHostf(host, format string, args ...any) {
	Default().logWithHost(LevelWarn, host, format, args...)
}

func ErrorHostf(host, format string, args ...any) {
	Default().logWithHost(LevelError, host, format, args...)
}

func Flush() {
	l := Default()
	l.mu.Lock()
	defer l.mu.Unlock()
}

func (l *Logger) log(level Level, format string, args ...any) {
	if level < l.level {
		return
	}
	formattedMsg := fmt.Sprintf(format, args...)
	coloredLevel := level.ColorString()
	logLine := fmt.Sprintf("%s %s\n", coloredLevel, formattedMsg)
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.out.Write([]byte(logLine))
}

func (l *Logger) logWithHost(level Level, host string, format string, args ...any) {
	formattedMsg := fmt.Sprintf(format, args...)
	hostPart := fmt.Sprintf("[%s]", blueColor(host))
	lineWithHost := fmt.Sprintf("%s %s", hostPart, formattedMsg)
	l.log(level, lineWithHost)
}
