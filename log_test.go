package log

// unit case for log
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
)

var (
	testLogFile = filepath.Join(os.TempDir(), "test.log")
)

func TestLogInit(t *testing.T) {
	SWLogger.Init(testLogFile, logrus.DebugLevel, true)
}

func TestIsLevelEnabled(t *testing.T) {
	SWLogger.Init(testLogFile, logrus.DebugLevel, true)
	if !SWLogger.isLevelEnabled(logrus.DebugLevel) {
		t.Error("isLevelEnabled failed")
	}
}
func TestLogf(t *testing.T) {
	SWLogger.Init(testLogFile, logrus.DebugLevel, true)
	SWLogger.Logf(logrus.DebugLevel, testLogFile, "test logf %s", "test")
}
func TestLog(t *testing.T) {
	SWLogger.Init(testLogFile, logrus.DebugLevel, true)
	SWLogger.Log(logrus.DebugLevel, testLogFile, "test log %s", "test")
}
func TestFormatterDecorator(t *testing.T) {
	SWLogger.Init(testLogFile, logrus.DebugLevel, true)
	SWLogger.formatterDecorator(testLogFile, "test")
}
func TestDebug(t *testing.T) {
	SWLogger.Init(testLogFile, logrus.DebugLevel, true)
	Debug("test debug")
}
func TestInfo(t *testing.T) {
	SWLogger.Init(testLogFile, logrus.DebugLevel, true)
	Info("test info")
}
func TestWarn(t *testing.T) {
	SWLogger.Init(testLogFile, logrus.DebugLevel, true)
	Warn("test warn")
}
func TestError(t *testing.T) {
	SWLogger.Init(testLogFile, logrus.DebugLevel, true)
	Error("test error")
}

func TestFatal(t *testing.T) {
	if os.Getenv("CRASH") == "1" {
		Fatal("Fatal test")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestFatal")
	cmd.Env = append(os.Environ(), "CRASH=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && e.Success() {
		t.Fatal("process ran with success which is not expected")
	}
}

func TestPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	Panic("Panic test")
}

func TestDebugf(t *testing.T) {
	SWLogger.Init(testLogFile, logrus.DebugLevel, true)
	Debugf("test debugf %s", "test")
}

func TestInfof(t *testing.T) {
	SWLogger.Init(testLogFile, logrus.DebugLevel, true)
	Infof("test infof %s", "test")
}

func TestErrorf(t *testing.T) {
	SWLogger.Init(testLogFile, logrus.DebugLevel, true)
	Errorf("test errorf %s", "test")
}

func TestWarnf(t *testing.T) {
	SWLogger.Init(testLogFile, logrus.DebugLevel, true)
	Warnf("test warnf %s", "test")
}

func TestFatalf(t *testing.T) {
	if os.Getenv("CRASH") == "1" {
		Fatalf("%s test", "Fatalf")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestFatalf")
	cmd.Env = append(os.Environ(), "CRASH=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && e.Success() {
		t.Fatal("process ran with success which is not expected")
	}
}

func TestPanicf(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	Panicf("%s test", "Panicf")
}

func TestSetLogLevel(t *testing.T) {
	SWLogger.Init(testLogFile, logrus.DebugLevel, true)
	SetLogLevel(logrus.InfoLevel)
}

func TestGetLogLevel(t *testing.T) {
	SWLogger.Init(testLogFile, logrus.DebugLevel, true)
	SetLogLevel(logrus.InfoLevel)
	level := GetLogLevel()
	if level != logrus.InfoLevel {
		t.Fatal("GetLogLevel failed")
	}
}
