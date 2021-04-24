package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
	"time"
)

var gLogger *Logger

type Logger struct {
	// origin logger by zap
	baseLogger *zap.Logger
	baseSugar  *zap.SugaredLogger

	filePath string
	logTime time.Time

	fileLevel       zapcore.Level
	consoleLevel    zapcore.Level
	stackTraceLevel zapcore.Level
}

func InitLog(fp string, fl string, cl string, stl string) {
	gLogger = &Logger{
		filePath: fp,
		fileLevel: parseLevel(fl),
		consoleLevel: parseLevel(cl),
		stackTraceLevel: parseLevel(stl),
	}

	gLogger.Reset()
}

func parseLevel(s string) zapcore.Level {
	switch strings.ToUpper(s) {
	case "DEBUG": return zapcore.DebugLevel
	case "INFO": return zapcore.InfoLevel
	case "WARN": return zapcore.WarnLevel
	case "ERROR": return zapcore.ErrorLevel
	case "FATAL": return zapcore.FatalLevel
	default: return zapcore.DebugLevel - 1
	}
}

func isPathExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if err == nil {
		return true
	}

	return os.IsExist(err)
}


func newWriteSync(filePath string) (zapcore.WriteSyncer, error) {
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

	return zapcore.AddSync(f), err
}

func newEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func NewZapLogger(filePath string, fileLevel zapcore.Level, consoleLevel zapcore.Level, stackTraceLevel zapcore.Level) *zap.Logger {
	writeSync, err := newWriteSync(filePath)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	fileCore := zapcore.NewCore(newEncoder(), writeSync, fileLevel)

	var logger *zap.Logger
	if consoleLevel >= zapcore.DebugLevel {
		consoleCore := zapcore.NewCore(newEncoder(), os.Stdout, consoleLevel)
		logger = zap.New(zapcore.NewTee(fileCore, consoleCore), zap.AddCaller(), zap.AddStacktrace(stackTraceLevel))

	} else {
		logger = zap.New(fileCore, zap.AddCaller(), zap.AddStacktrace(stackTraceLevel))
	}

	return logger
}

func (this *Logger) Close() {
	this.baseLogger.Sync()
	this.baseSugar = nil
	this.baseLogger = nil
}

func (this *Logger) Reset() {
	this.baseLogger = NewZapLogger(this.filePath, this.fileLevel, this.consoleLevel, this.stackTraceLevel)
	this.baseSugar = this.baseLogger.Sugar()
	this.logTime = time.Now()
}

func (this *Logger) checkTime() {
	if this.logTime.Day() != time.Now().Day() {
		this.baseLogger.Sync()
		this.Reset()
	}
}

func (this *Logger) Debug(template string, args ...interface{}) {
	this.checkTime()
	this.baseSugar.Debugf(template, args...)
}

func (this *Logger) Info(template string, args ...interface{}) {
	this.checkTime()
	this.baseSugar.Infof(template, args...)
}

func (this *Logger) Warn(template string, args ...interface{}) {
	this.checkTime()
	this.baseSugar.Warnf(template, args...)
}

func (this *Logger) Error(template string, args ...interface{}) {
	this.checkTime()
	this.baseSugar.Errorf(template, args...)
}

func (this *Logger) Fatal(template string, args ...interface{}) {
	this.checkTime()
	this.baseSugar.Fatalf(template, args...)
}

func Debug(template string, args ...interface{}) {
	gLogger.Debug(template, args...)
}

func Info(template string, args ...interface{}) {
	gLogger.Info(template, args...)
}

func Warn(template string, args ...interface{}) {
	gLogger.Warn(template, args...)
}

func Error(template string, args ...interface{}) {
	gLogger.Error(template, args...)
}

func Fatal(template string, args ...interface{}) {
	gLogger.Fatal(template, args...)
}