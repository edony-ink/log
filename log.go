// Package log is a well formatted golang logging library.
package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

const (
	// Default log format will output:
	// `|2006-01-02 15:04:05,123|INFO   |func(path/xxx.go:line)|Log message|`
	defaultLogFormat       = "%time%|%lvl%|%func%(%line%)|%msg%\n"
	defaultTimestampFormat = time.StampMilli

	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel = logrus.PanicLevel
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel = logrus.FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel = logrus.ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel = logrus.WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel = logrus.InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel = logrus.DebugLevel

	// DayHours is the hours of day
	DayHours = 24
	// WeekHours is the hours of week
	WeekHours = 7 * 24
)

// SWLog wrap the log system for
type SWLog struct {
	// log will output log into log file and stderr
	STDLogger  *logrus.Logger
	FileLogger *logrus.Logger
	IsLog2STD  bool
	LogFile    string
	LogLevel   logrus.Level
	// skip is the number of stack frames to ascend
	skip int
	// isSetup is to make sure SWLog has been setup
	isSetup bool
	// raw logging without format
	raw bool
}

// Formatter implements logrus.Formatter interface.
type Formatter struct {
	// Timestamp format
	logrus.TextFormatter
	// Available standard keys: time, msg, lvl
	// Also can include custom fields but limited to strings.
	// All of fields need to be wrapped inside %% i.e %time% %msg%
	LogFormat string
	// file name and line number where calling the LOG/INFO/DEBUG...
	FileName string
	// function name where calling the LOG/INFO/DEBUG...
	FuncName string
}

// Format building log message.
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	output := f.LogFormat
	if output == "" {
		output = defaultLogFormat
	}

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = defaultTimestampFormat
	}
	currentDate := strings.Split(entry.Time.Format(time.RFC3339), "T")
	if len(currentDate) != 2 {
		return nil, fmt.Errorf("split date format error")
	}
	currentTime := strings.Split(entry.Time.Format(timestampFormat), " ")
	if len(currentTime) < 1 {
		return nil, fmt.Errorf("split time format error")
	}

	output = strings.Replace(output, "%time%", currentDate[0]+" "+currentTime[len(currentTime)-1], 1)
	output = strings.Replace(output, "%msg%", entry.Message, 1)

	level := strings.ToUpper(entry.Level.String())
	// keep log level info left-justifying
	if len(level) < 7 {
		for i := 7 - len(level); i > 0; i-- {
			level += " "
		}
	}
	output = strings.Replace(output, "%lvl%", level, 1)

	output = strings.Replace(output, "%line%", f.FileName, 1)
	output = strings.Replace(output, "%func%", f.FuncName, 1)

	for k, v := range entry.Data {
		if s, ok := v.(string); ok {
			output = strings.Replace(output, "%"+k+"%", s, 1)
		}
	}

	return []byte(output), nil
}

type STDFormatter struct {
	LogFormat string
	// file name and line number where calling the LOG/INFO/DEBUG...
	FileName string
	// function name where calling the LOG/INFO/DEBUG...
	FuncName string
}

func (f *STDFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	output := "%msg%\n"
	output = strings.Replace(output, "%msg%", entry.Message, 1)

	return []byte(output), nil
}

// Init setup SWLogger before running
func (logger *SWLog) Init(logFile string, level logrus.Level, log2STD bool) {
	if logger.isSetup {
		logger.Debug("no need to setup, swlogger is already setup!")
		return
	}

	// init log level
	logger.LogLevel = level

	// init log directory and log file
	if logFile != "" {
		// SWLogger log into `defaultLogFile` by default
		// change the log file with parameter logFile
		logger.LogFile = logFile
	}
	if logger.LogFile == "" {
		logrus.Fatalf("Log file or directory is not set")
	}
	// monitor requires the permission of the log directory as `o+rx' and log file as `o+r'
	dir := filepath.Dir(logger.LogFile)
	if _, err := os.Stat(dir); err != nil {
		if !os.IsExist(err) {
			err := os.MkdirAll(dir, 0777)
			if err != nil {
				logrus.Fatalf("Fail to mkdir %s; %s", dir, err.Error())
			}
		}
	}

	// init file logrus logger
	logger.FileLogger = &logrus.Logger{
		Level:     logger.LogLevel,
		Hooks:     make(logrus.LevelHooks),
		Formatter: &Formatter{},
	}

	// init logrotate
	writer, err := rotatelogs.New(
		logger.LogFile+".%Y-%m-%d",
		// create a new Option that sets the symbolic link name that gets linked to the current file name being used.
		rotatelogs.WithLinkName(logger.LogFile),

		// create a new Option that sets the time between rotation(default 24 hours).
		rotatelogs.WithRotationTime(DayHours*time.Hour),

		// create a new Option that sets the number of files should be kept before it gets purged from the file system.
		//rotatelogs.WithRotationCount(1000),
		// creates a new Option that sets the max age of a log file before it gets purged from the file system.
		rotatelogs.WithMaxAge(WeekHours*time.Hour),

		// max rotated file size is 512Mb
		rotatelogs.WithRotationSize(512*1024),
	)
	if err != nil {
		logrus.Fatalf("config local file system for logger error: %s", err.Error())
	}
	logger.FileLogger.SetOutput(writer)

	// init log stack frames to ascend
	logger.skip = 2

	// init IsLog2STD
	logger.IsLog2STD = log2STD

	// init logrus.logger
	logger.STDLogger = &logrus.Logger{
		Out:       os.Stderr,
		Level:     SWLogger.LogLevel,
		Formatter: &Formatter{},
	}

	// init finished
	logger.isSetup = true
}

func (logger *SWLog) SetRawSTDLogging(isRaw bool) {
	logger.raw = isRaw
	if logger.raw {
		logger.STDLogger.SetFormatter(&STDFormatter{})
	} else {
		logger.STDLogger.SetFormatter(&Formatter{})
	}
}

func (logger *SWLog) formatterDecorator(filename string, funcname string) {
	if logger.raw {
		(logger.STDLogger.Formatter).(*STDFormatter).FileName = filename
		(logger.STDLogger.Formatter).(*STDFormatter).FuncName = funcname
	} else {
		(logger.STDLogger.Formatter).(*Formatter).FileName = filename
		(logger.STDLogger.Formatter).(*Formatter).FuncName = funcname
	}
	(logger.FileLogger.Formatter).(*Formatter).FileName = filename
	(logger.FileLogger.Formatter).(*Formatter).FuncName = funcname
}

func (logger *SWLog) isLevelEnabled(level logrus.Level) bool {
	return logger.LogLevel >= level
}

// Log logging the message of input args...
func (logger *SWLog) Log(level logrus.Level, filename string, funcname string, args ...interface{}) {
	// SWLog needs to be setup first before logging
	if !logger.isSetup {
		logrus.Fatal("log not setup which will cause panic")
	}

	// decorate log with calling info
	logger.formatterDecorator(filename, funcname)

	if logger.isLevelEnabled(level) {
		switch level {
		case logrus.PanicLevel:
			if logger.IsLog2STD {
				logger.STDLogger.Panic(args...)
			}
			logger.FileLogger.Panic(args...)
		case logrus.FatalLevel:
			if logger.IsLog2STD {
				logger.STDLogger.Fatal(args...)
			}
			logger.FileLogger.Fatal(args...)
		case logrus.ErrorLevel:
			if logger.IsLog2STD {
				logger.STDLogger.Error(args...)
			}
			logger.FileLogger.Error(args...)
		case logrus.WarnLevel:
			if logger.IsLog2STD {
				logger.STDLogger.Warn(args...)
			}
			logger.FileLogger.Warn(args...)
		case logrus.InfoLevel:
			if logger.IsLog2STD {
				logger.STDLogger.Info(args...)
			}
			logger.FileLogger.Info(args...)
		default:
			if logger.IsLog2STD {
				logger.STDLogger.Debug(args...)
			}
			logger.FileLogger.Debug(args...)
		}
	}
}

// Logf logging the message with the given formated args
func (logger *SWLog) Logf(level logrus.Level, format string, filename string, funcname string, args ...interface{}) {
	logger.Log(level, filename, funcname, fmt.Sprintf(format, args...))
}

// Debug logging in debug level
func (logger *SWLog) Debug(args ...interface{}) {
	var dir, filename, funcname string
	var line int
	pc, filename, line, ok := runtime.Caller(logger.skip)
	if ok {
		funcname = runtime.FuncForPC(pc).Name()      // main.(*MyStruct).foo
		funcname = filepath.Ext(funcname)            // .foo
		funcname = strings.TrimPrefix(funcname, ".") // foo

		dir, filename = filepath.Split(filename)
		filename = filepath.Base(dir) + "/" + filepath.Base(filename) + ":" + strconv.FormatInt(int64(line), 10) // /full/path/basename.go => basename.go
	}

	logger.Log(logrus.DebugLevel, filename, funcname, args...)
}

// Info logging in info level
func (logger *SWLog) Info(args ...interface{}) {
	var dir, filename, funcname string
	var line int
	pc, filename, line, ok := runtime.Caller(logger.skip)
	if ok {
		funcname = runtime.FuncForPC(pc).Name()      // main.(*MyStruct).foo
		funcname = filepath.Ext(funcname)            // .foo
		funcname = strings.TrimPrefix(funcname, ".") // foo

		dir, filename = filepath.Split(filename)
		filename = filepath.Base(dir) + "/" + filepath.Base(filename) + ":" + strconv.FormatInt(int64(line), 10) // /full/path/basename.go => basename.go
	}

	logger.Log(logrus.InfoLevel, filename, funcname, args...)
}

// Warn logging in warning level
func (logger *SWLog) Warn(args ...interface{}) {
	var dir, filename, funcname string
	var line int
	pc, filename, line, ok := runtime.Caller(logger.skip)
	if ok {
		funcname = runtime.FuncForPC(pc).Name()      // main.(*MyStruct).foo
		funcname = filepath.Ext(funcname)            // .foo
		funcname = strings.TrimPrefix(funcname, ".") // foo

		dir, filename = filepath.Split(filename)
		filename = filepath.Base(dir) + "/" + filepath.Base(filename) + ":" + strconv.FormatInt(int64(line), 10) // /full/path/basename.go => basename.go
	}

	logger.Log(logrus.WarnLevel, filename, funcname, args...)
}

// Error logging in error level
func (logger *SWLog) Error(args ...interface{}) {
	var dir, filename, funcname string
	var line int
	pc, filename, line, ok := runtime.Caller(logger.skip)
	if ok {
		funcname = runtime.FuncForPC(pc).Name()      // main.(*MyStruct).foo
		funcname = filepath.Ext(funcname)            // .foo
		funcname = strings.TrimPrefix(funcname, ".") // foo

		dir, filename = filepath.Split(filename)
		filename = filepath.Base(dir) + "/" + filepath.Base(filename) + ":" + strconv.FormatInt(int64(line), 10) // /full/path/basename.go => basename.go
	}

	logger.Log(logrus.ErrorLevel, filename, funcname, args...)
}

// Fatal logging in fatal level and calling the os.Exit()
func (logger *SWLog) Fatal(args ...interface{}) {
	var dir, filename, funcname string
	var line int
	pc, filename, line, ok := runtime.Caller(logger.skip)
	if ok {
		funcname = runtime.FuncForPC(pc).Name()      // main.(*MyStruct).foo
		funcname = filepath.Ext(funcname)            // .foo
		funcname = strings.TrimPrefix(funcname, ".") // foo

		dir, filename = filepath.Split(filename)
		filename = filepath.Base(dir) + "/" + filepath.Base(filename) + ":" + strconv.FormatInt(int64(line), 10) // /full/path/basename.go => basename.go
	}

	logger.Log(logrus.FatalLevel, filename, funcname, args...)
}

// Panic logging in panic level and calling the os.Panic()
func (logger *SWLog) Panic(args ...interface{}) {
	var dir, filename, funcname string
	var line int
	pc, filename, line, ok := runtime.Caller(logger.skip)
	if ok {
		funcname = runtime.FuncForPC(pc).Name()      // main.(*MyStruct).foo
		funcname = filepath.Ext(funcname)            // .foo
		funcname = strings.TrimPrefix(funcname, ".") // foo

		dir, filename = filepath.Split(filename)
		filename = filepath.Base(dir) + "/" + filepath.Base(filename) + ":" + strconv.FormatInt(int64(line), 10) // /full/path/basename.go => basename.go
	}

	logger.Log(logrus.PanicLevel, filename, funcname, args...)
}

// Debugf logging in debug level with the given formated args
func (logger *SWLog) Debugf(format string, args ...interface{}) {
	var dir, filename, funcname string
	var line int
	pc, filename, line, ok := runtime.Caller(logger.skip)
	if ok {
		funcname = runtime.FuncForPC(pc).Name()      // main.(*MyStruct).foo
		funcname = filepath.Ext(funcname)            // .foo
		funcname = strings.TrimPrefix(funcname, ".") // foo

		dir, filename = filepath.Split(filename)
		filename = filepath.Base(dir) + "/" + filepath.Base(filename) + ":" + strconv.FormatInt(int64(line), 10) // /full/path/basename.go => basename.go
	}

	logger.Logf(logrus.DebugLevel, format, filename, funcname, args...)
}

// Infof logging in info level
func (logger *SWLog) Infof(format string, args ...interface{}) {
	var dir, filename, funcname string
	var line int
	pc, filename, line, ok := runtime.Caller(logger.skip)
	if ok {
		funcname = runtime.FuncForPC(pc).Name()      // main.(*MyStruct).foo
		funcname = filepath.Ext(funcname)            // .foo
		funcname = strings.TrimPrefix(funcname, ".") // foo

		dir, filename = filepath.Split(filename)
		filename = filepath.Base(dir) + "/" + filepath.Base(filename) + ":" + strconv.FormatInt(int64(line), 10) // /full/path/basename.go => basename.go
	}
	logger.Logf(logrus.InfoLevel, format, filename, funcname, args...)
}

// Warnf logging in warning level
func (logger *SWLog) Warnf(format string, args ...interface{}) {
	var dir, filename, funcname string
	var line int
	pc, filename, line, ok := runtime.Caller(logger.skip)
	if ok {
		funcname = runtime.FuncForPC(pc).Name()      // main.(*MyStruct).foo
		funcname = filepath.Ext(funcname)            // .foo
		funcname = strings.TrimPrefix(funcname, ".") // foo

		dir, filename = filepath.Split(filename)
		filename = filepath.Base(dir) + "/" + filepath.Base(filename) + ":" + strconv.FormatInt(int64(line), 10) // /full/path/basename.go => basename.go
	}
	logger.Logf(logrus.WarnLevel, format, filename, funcname, args...)
}

// Errorf logging in error level
func (logger *SWLog) Errorf(format string, args ...interface{}) {
	var dir, filename, funcname string
	var line int
	pc, filename, line, ok := runtime.Caller(logger.skip)
	if ok {
		funcname = runtime.FuncForPC(pc).Name()      // main.(*MyStruct).foo
		funcname = filepath.Ext(funcname)            // .foo
		funcname = strings.TrimPrefix(funcname, ".") // foo

		dir, filename = filepath.Split(filename)
		filename = filepath.Base(dir) + "/" + filepath.Base(filename) + ":" + strconv.FormatInt(int64(line), 10) // /full/path/basename.go => basename.go
	}
	logger.Logf(logrus.ErrorLevel, format, filename, funcname, args...)
}

// Fatalf logging in fatal level and calling the os.Exit()
func (logger *SWLog) Fatalf(format string, args ...interface{}) {
	var dir, filename, funcname string
	var line int
	pc, filename, line, ok := runtime.Caller(logger.skip)
	if ok {
		funcname = runtime.FuncForPC(pc).Name()      // main.(*MyStruct).foo
		funcname = filepath.Ext(funcname)            // .foo
		funcname = strings.TrimPrefix(funcname, ".") // foo

		dir, filename = filepath.Split(filename)
		filename = filepath.Base(dir) + "/" + filepath.Base(filename) + ":" + strconv.FormatInt(int64(line), 10) // /full/path/basename.go => basename.go
	}
	logger.Logf(logrus.FatalLevel, format, filename, funcname, args...)
}

// Panicf logging in panic level and calling the os.Panic()
func (logger *SWLog) Panicf(format string, args ...interface{}) {
	var dir, filename, funcname string
	var line int
	pc, filename, line, ok := runtime.Caller(logger.skip)
	if ok {
		funcname = runtime.FuncForPC(pc).Name()      // main.(*MyStruct).foo
		funcname = filepath.Ext(funcname)            // .foo
		funcname = strings.TrimPrefix(funcname, ".") // foo

		dir, filename = filepath.Split(filename)
		filename = filepath.Base(dir) + "/" + filepath.Base(filename) + ":" + strconv.FormatInt(int64(line), 10) // /full/path/basename.go => basename.go
	}
	logger.Logf(logrus.PanicLevel, format, filename, funcname, args...)
}

/*
 * SWLogger is the global logging instance for out of box logging
 */
const (
	DefaultLogDir     = "/tmp/log/"
	DefaultLogFile    = "log.log"
	DefaultCLILogFile = "cli.log"
)

var (
	// SWLogger is global logging instance, which need to call `SWLog.Init()` to setup before running
	SWLogger = &SWLog{
		isSetup: false,
		raw:     false,
	}

	// LevelFromStr map configuration string with log level
	LevelFromStr = map[string]logrus.Level{
		"PANIC": PanicLevel,
		"FATAL": FatalLevel,
		"ERROR": ErrorLevel,
		"WARN":  WarnLevel,
		"INFO":  InfoLevel,
		"DEBUG": DebugLevel,
	}
)

// SetLogLevel set the log level
func SetLogLevel(level logrus.Level) {
	if SWLogger.IsLog2STD {
		SWLogger.STDLogger.SetLevel(level)
	}

	SWLogger.FileLogger.SetLevel(level)
	SWLogger.LogLevel = level
}

// GetLogLevel get the log level
func GetLogLevel() logrus.Level {
	// loggerSTD and loggerF got the same LogLevel
	return SWLogger.FileLogger.GetLevel()
}

// Debug logging in debug level
func Debug(args ...interface{}) {
	SWLogger.Debug(args...)
}

// Info logging in info level
func Info(args ...interface{}) {
	SWLogger.Info(args...)
}

// Warn logging in warning level
func Warn(args ...interface{}) {
	SWLogger.Warn(args...)
}

// Error logging in error level
func Error(args ...interface{}) {
	SWLogger.Error(args...)
}

// Fatal logging in fatal level and calling the os.Exit()
func Fatal(args ...interface{}) {
	SWLogger.Fatal(args...)
}

// Panic logging in panic level and calling the os.Panic()
func Panic(args ...interface{}) {
	SWLogger.Panic(args...)
}

// Debugf logging in debug level with the given formated args
func Debugf(format string, args ...interface{}) {
	SWLogger.Debugf(format, args...)
}

// Infof logging in info level
func Infof(format string, args ...interface{}) {
	SWLogger.Infof(format, args...)
}

// Warnf logging in warning level
func Warnf(format string, args ...interface{}) {
	SWLogger.Warnf(format, args...)
}

// Errorf logging in error level
func Errorf(format string, args ...interface{}) {
	SWLogger.Errorf(format, args...)
}

// Fatalf logging in fatal level and calling the os.Exit()
func Fatalf(format string, args ...interface{}) {
	SWLogger.Fatalf(format, args...)
}

// Panicf logging in panic level and calling the os.Panic()
func Panicf(format string, args ...interface{}) {
	SWLogger.Panicf(format, args...)
}
