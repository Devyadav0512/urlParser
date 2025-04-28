package utils

import (
	"log"
	"os"
)

type Logger struct {
	infoLog  *log.Logger
	warnLog  *log.Logger  // Added warning logger
	errorLog *log.Logger
	debugLog *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		infoLog:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLog:  log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),  // Added
		errorLog: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLog: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *Logger) Info(message string, args ...interface{}) {
	l.infoLog.Printf(message, args...)
}

// Added Warn method
func (l *Logger) Warn(message string, args ...interface{}) {
	l.warnLog.Printf(message, args...)
}

func (l *Logger) Error(message string, args ...interface{}) {
	l.errorLog.Printf(message, args...)
}

func (l *Logger) Debug(message string, args ...interface{}) {
	l.debugLog.Printf(message, args...)
}