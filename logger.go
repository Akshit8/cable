package cable

import (
	"log"
	"os"
)

// Logger interface
type Logger interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

// NewLogger return the default logger
func NewLogger() Logger {
	return newDefaultLogger()
}

// default logger uses standard log package to implement logger interface
type defaultLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger
}

func newDefaultLogger() Logger {
	return &defaultLogger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime),
		fatalLogger: log.New(os.Stderr, "FATAL: ", log.Ldate|log.Ltime),
	}
}

func (d *defaultLogger) Info(args ...interface{}) {
	d.infoLogger.Println(args...)
}

func (d *defaultLogger) Infof(format string, args ...interface{}) {
	d.infoLogger.Printf(format, args...)
}

func (d *defaultLogger) Error(args ...interface{}) {
	d.errorLogger.Println(args...)
}

func (d *defaultLogger) Errorf(format string, args ...interface{}) {
	d.errorLogger.Printf(format, args...)
}

func (d *defaultLogger) Fatal(args ...interface{}) {
	d.fatalLogger.Fatalln(args...)
}

func (d *defaultLogger) Fatalf(format string, args ...interface{}) {
	d.fatalLogger.Fatalf(format, args...)
}
