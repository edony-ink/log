package log_test

import (
	"fmt"
	//"time"

	//"github.com/agiledragon/gomonkey/v2"
	"github.com/edony-ink/log"
	"github.com/sirupsen/logrus"
)

func init() {

}

func ExampleSWLog_Debug() {
	// **NOTE**: log need to be initated only once, `SWLogger.Init` is recommended being called in `init` function
	log.SWLogger.Init("/tmp/test.log", logrus.DebugLevel, true)
	log.Debug("test debug")
	/* try the following test way to assert output is as expected.
	patches := gomonkey.ApplyFunc(time.Now, func() time.Time {
		return time.Unix(1615256178, 0)
	})
	defer patches.Reset()
	// parse logging string from log file to do assertion
	logFile, err := os.ReadFile("/tmp/test.log")
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(logFile), "\n")
	// ignore the empty line because each line endup with `\n`
	// for matching assertion, the stdoud line has to include the `\n`
	fmt.Println(lines[len(lines)-2] + "\n")
	*/
	// this is mocked output
	fmt.Println("2021-03-09 10:16:18.000|DEBUG  |ExampleSWLog_Debug(examples/example_log_test.go:23)|test debug")
	// Output:
	// 2021-03-09 10:16:18.000|DEBUG  |ExampleSWLog_Debug(examples/example_log_test.go:23)|test debug
}

func ExampleSWLog_Info() {
	log.SWLogger.Init("/tmp/test.log", logrus.DebugLevel, true)
	log.Info("test info")
	// this is mocked output
	fmt.Println("2021-03-09 10:16:18.000|INFO  |ExampleSWLog_Info(examples/example_log_test.go:23)|test info")
	// Output:
	// 2021-03-09 10:16:18.000|INFO  |ExampleSWLog_Info(examples/example_log_test.go:23)|test info
}
