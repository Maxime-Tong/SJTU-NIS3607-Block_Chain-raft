package mylogger

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type MyLogger struct {
	Debug  bool
	logger *log.Logger
}

func InitLogger(name string, id uint8) *MyLogger {
	logfile := "../log/" + name + strconv.Itoa(int(id)) + ".log"
	f, err := os.OpenFile(logfile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		fmt.Println("create log file failed:", err.Error())
		return nil
	}
	logger := log.New(f, "[Consesus log] ", log.LstdFlags|log.Lshortfile|log.LUTC|log.Lmicroseconds)
	mylogger := &MyLogger{
		Debug:  true,
		logger: logger,
	}
	return mylogger
}

func (l *MyLogger) DPrintf(format string, a ...interface{}) (err error) {
	if l.Debug {
		l.logger.Printf(format, a...)
	}
	return nil
}
