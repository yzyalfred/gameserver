package log

import (
	"fmt"
	"gameserver/core/utils"
	"log"
	"os"
	"strings"
	"time"
)

// levels
const (
	debugLevel = iota
	infoLevel
	warnLevel
	errorLevel
	fatalLevel
)

const (
	debugLevelString = "[DEBUG] "
	infoLevelString  = "[INFO ] "
	warnLevelString  = "[WARN] "
	errorLevelString = "[ERROR] "
	fatalLevelString = "[FATAL] "
)

var gLogger *Logger

type Logger struct {
	fileLogger *log.Logger
	stdLogger  *log.Logger

	filePath string
	file     *os.File

	logTime time.Time
	flag    int
	level   int
}

func InitLog(filePath string, level string, std bool, flag int) {
	if flag == 0 {
		flag = log.LstdFlags | log.Lshortfile
	}

	gLogger = &Logger{
		filePath: filePath,
		level:    parseLevel(level),
		logTime:  time.Now(),
		flag:     flag,
	}

	file, err := OpenOrCreateFile(gLogger.filePath)
	if err != nil {
		fmt.Println("file to new logger", err)
		return
	}

	gLogger.fileLogger = log.New(file, "", gLogger.flag)
	gLogger.file = file
	if std {
		gLogger.stdLogger = log.New(os.Stdout, "", gLogger.flag)
	}
}

func parseLevel(s string) int {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return debugLevel
	case "INFO":
		return infoLevel
	case "WARN":
		return warnLevel
	case "ERROR":
		return errorLevel
	case "FATAL":
		return fatalLevel
	default:
		return infoLevel
	}
}

func isPathExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if err == nil {
		return true
	}

	return os.IsExist(err)
}

func OpenOrCreateFile(filePath string) (*os.File, error) {
	now := time.Now()
	fileName := fmt.Sprintf("%s/%d-%02d-%02d.log",
		filePath,
		now.Year(),
		now.Month(),
		now.Day(),
	)

	var f *os.File
	var err error
	if isPathExist(fileName) {
		f, err = os.OpenFile(fileName, os.O_APPEND, 0)

	} else {
		f, err = os.Create(fileName)
	}

	return f, err
}

func (this *Logger) Close() {

}

func (this *Logger) checkTime() {
	now := time.Now()
	if this.logTime.Day() != now.Day() {
		file, err := OpenOrCreateFile(this.filePath)
		if err == nil {
			this.file.Close()
			this.fileLogger = log.New(file, "", this.flag)
			this.file = file
			this.logTime = now
		} else {
			// log error and use old file
			this.Error("create file error: ", err)
		}

	}
}

func (this *Logger) doPrint(level int, levelString string, format string, args ...interface{}) {
	if level < this.level {
		return
	}

	format = levelString +  format
	this.checkTime()
	if level >= errorLevel {
		format = fmt.Sprintf(format, args...)
		format = format + "\r\n" + utils.TakeStacktrace(3)

	} else {
		format = fmt.Sprintf(format, args...)
	}

	this.fileLogger.Output(3, format)
	if this.stdLogger != nil {
		this.stdLogger.Output(3, format)
	}

	if level == fatalLevel {
		os.Exit(1)
	}
}

func (this *Logger) Debug(format string, args ...interface{}) {
	this.checkTime()
	this.doPrint(debugLevel, debugLevelString, "", args...)
}

func (this *Logger) Info(format string, args ...interface{}) {
	this.checkTime()
	this.doPrint(infoLevel, infoLevelString, format, args...)
}

func (this *Logger) Warn(format string, args ...interface{}) {
	this.checkTime()
	this.doPrint(warnLevel, warnLevelString, format, args...)
}

func (this *Logger) Error(format string, args ...interface{}) {
	this.checkTime()
	this.doPrint(errorLevel, errorLevelString, format, args...)
}

func (this *Logger) Fatal(format string, args ...interface{}) {
	this.checkTime()
	this.doPrint(fatalLevel, fatalLevelString, format, args...)
}

func Debug(format string, args ...interface{}) {
	gLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	gLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	gLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	gLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	gLogger.Fatal(format, args...)
}
