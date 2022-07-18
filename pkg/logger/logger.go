package logger

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	log *logrus.Entry
	fd  *os.File
}

func (l *Logger) Debug(args ...interface{}) {
	if l.log != nil {
		l.log.Log(logrus.DebugLevel, args...)
	}
}

func (l *Logger) Info(args ...interface{}) {
	if l.log != nil {
		l.log.Log(logrus.InfoLevel, args...)
	}
}

func (l *Logger) LogOnErr(err error) {
	if err != nil {
		l.Error(err)
	}
}

func (l *Logger) Error(args ...interface{}) {
	if l.log != nil {
		l.log.Log(logrus.ErrorLevel, args...)
	}
}

func (l *Logger) Fatal(args ...interface{}) {
	if l.log != nil {
		l.log.Fatal(args...)
	}

	log.Fatal(args...)
}

func (l *Logger) Close() {
	if l.fd == nil {
		return
	}

	if err := l.fd.Close(); err != nil {
		log.Println("failed to close file:", err)
	}
}

func Init(file string) *Logger {
	if len(file) == 0 {
		return &Logger{}
	}

	fd, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_RDWR, os.FileMode(0600))
	if err != nil {
		return &Logger{}
	}

	logger := logrus.New()

	logger.Formatter = &logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			fn := path.Base(frame.Function)

			return fmt.Sprintf("%s()", frame.Function), fmt.Sprintf(" %s:%d", fn, frame.Line)
		},
	}

	logger.SetOutput(fd)

	return &Logger{logrus.NewEntry(logger), fd}
}
