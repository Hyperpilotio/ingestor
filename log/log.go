package log

import (
	"os"
	"path/filepath"

	"github.com/hyperpilotio/ingestor/config"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

type Logger = logrus.Logger
type LoggerEntry = logrus.Entry

var defaultLogger *Logger

func init() {
	defaultLogger = NewLogger(*config.Config())
}

func NewLogger(cfg config.Provider) *Logger {
	l := logrus.New()

	if cfg.GetBool("JSONLog") {
		l.Formatter = &logrus.JSONFormatter{}
	}

	l.Out = os.Stderr
	if cfg.GetBool("LogFile") {
		l.Hooks.Add(lfshook.NewHook(lfshook.PathMap{
			logrus.WarnLevel:  filepath.Join(cfg.GetString("LogFolder"), "ingestor-warn.log"),
			logrus.InfoLevel:  filepath.Join(cfg.GetString("LogFolder"), "ingestor-info.log"),
			logrus.ErrorLevel: filepath.Join(cfg.GetString("LogFolder"), "ingestor-error.log"),
			logrus.FatalLevel: filepath.Join(cfg.GetString("LogFolder"), "ingestor-fatal.log"),
		}))
	}

	switch cfg.GetString("LogLevel") {
	case "debug":
		l.Level = logrus.DebugLevel
	case "warning":
		l.Level = logrus.WarnLevel
	case "info":
		l.Level = logrus.InfoLevel
	default:
		l.Level = logrus.DebugLevel
	}

	return l
}

type Fields map[string]interface{}

func (f Fields) With(k string, v interface{}) Fields {
	f[k] = v
	return f
}

func (f Fields) WithFields(f2 Fields) Fields {
	for k, v := range f2 {
		f[k] = v
	}
	return f
}

func WithFields(fields Fields) *LoggerEntry {
	return defaultLogger.WithFields(logrus.Fields(fields))
}

// Debug package-level convenience method.
func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

// Debugf package-level convenience method.
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Debugln package-level convenience method.
func Debugln(args ...interface{}) {
	defaultLogger.Debugln(args...)
}

// Error package-level convenience method.
func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

// Errorf package-level convenience method.
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Errorln package-level convenience method.
func Errorln(args ...interface{}) {
	defaultLogger.Errorln(args...)
}

// Fatal package-level convenience method.
func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

// Fatalf package-level convenience method.
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

// Fatalln package-level convenience method.
func Fatalln(args ...interface{}) {
	defaultLogger.Fatalln(args...)
}

// Info package-level convenience method.
func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

// Infof package-level convenience method.
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Infoln package-level convenience method.
func Infoln(args ...interface{}) {
	defaultLogger.Infoln(args...)
}

// Panic package-level convenience method.
func Panic(args ...interface{}) {
	defaultLogger.Panic(args...)
}

// Panicf package-level convenience method.
func Panicf(format string, args ...interface{}) {
	defaultLogger.Panicf(format, args...)
}

// Panicln package-level convenience method.
func Panicln(args ...interface{}) {
	defaultLogger.Panicln(args...)
}

// Print package-level convenience method.
func Print(args ...interface{}) {
	defaultLogger.Print(args...)
}

// Printf package-level convenience method.
func Printf(format string, args ...interface{}) {
	defaultLogger.Printf(format, args...)
}

// Println package-level convenience method.
func Println(args ...interface{}) {
	defaultLogger.Println(args...)
}

// Warn package-level convenience method.
func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

// Warnf package-level convenience method.
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

// Warning package-level convenience method.
func Warning(args ...interface{}) {
	defaultLogger.Warning(args...)
}

// Warningf package-level convenience method.
func Warningf(format string, args ...interface{}) {
	defaultLogger.Warningf(format, args...)
}

// Warningln package-level convenience method.
func Warningln(args ...interface{}) {
	defaultLogger.Warningln(args...)
}

// Warnln package-level convenience method.
func Warnln(args ...interface{}) {
	defaultLogger.Warnln(args...)
}
