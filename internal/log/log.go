package log

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// Log defines the function signature of the logger
type Log func(l ...interface{})

// Log defines the function signature of the logger when called with format parameters
type Logf func(s string, l ...interface{})

// Trace logs at the Trace level
var Trace,
	// Debug logs at the Debug level
	Debug,
	// Info logs at the Info level
	Info,
	// Warn logs at the Warn level
	Warn,
	// Error logs at the Error level
	Error,
	Print Log = logger.Trace,
	logger.Debug,
	logger.Info,
	logger.Warn,
	logger.Error,
	func(l ...interface{}) {
		fmt.Println(l...)
	}

// Tracef logs at the trace level with formatting
var Tracef,
	// Debugf logs at the debug level with formatting
	Debugf,
	// Infof logs at the info level with formatting
	Infof,
	// Warnf logs at the warn level with formatting
	Warnf,
	// Errorf logs at the error level with formatting
	Errorf,
	Printf Logf = logger.Tracef,
	logger.Debugf,
	logger.Infof,
	logger.Warnf,
	logger.Errorf,
	func(s string, l ...interface{}) {
		fmt.Printf(s, l...)
		fmt.Printf("\n")
	}

func init() {
	logger.SetLevel(logrus.TraceLevel)

	logger.SetOutput(os.Stderr)
	logger.SetReportCaller(false)
	logger.SetFormatter(&logrus.TextFormatter{
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			function = path.Base(frame.Function)
			file = path.Base(frame.File)
			return
		},
		TimestampFormat:  "15:04:05",
		DisableSorting:   true,
		FullTimestamp:    true,
		QuoteEmptyFields: true,
	})
}
