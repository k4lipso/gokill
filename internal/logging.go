package internal

import (
	"github.com/apsdehal/go-logger"
	"fmt"
	"os"
)

var Log *logger.Logger
var documenterLoggers map[string]*logger.Logger
var verbose bool

var formatNormal string = "%{message}"
var formatVerbose string = "%{time} %{level}: %{message}"
var formatModuleNormal string = "%{module}: %{message}"
var formatModuleVerbose string = "%{time} %{level} %{module}: %{message}"

func InitLogger() error {
	var err error
	Log, err = logger.New("UsbDisconnect", 1, os.Stdout)

	if err != nil {
		return fmt.Errorf("Logger initialization error: %s", err)
	}

	documenterLoggers = make(map[string]*logger.Logger)
	verbose = false
	Log.SetLogLevel(logger.InfoLevel)

	return nil
}

func LogDoc(documenter Documenter) *logger.Logger {
	documenterLogger, ok := documenterLoggers[documenter.GetName()]

	if !ok {
		documenterLogger, _ = logger.New(documenter.GetName(), 1, os.Stdout)
		documenterLoggers[documenter.GetName()] = documenterLogger
	}

	SetVerbose(verbose)
	//documenterLogger.SetFormat("%{message}")

	return documenterLogger
}

func SetVerbose(value bool) {
	verbose = value
	toggleVerbosity := func(l *logger.Logger, formatN string, formatV string) {
			if value {
			l.SetLogLevel(logger.DebugLevel)
			l.SetFormat(formatV)
		} else {
			l.SetLogLevel(logger.InfoLevel)
			l.SetFormat(formatN)
		}
	}

	toggleVerbosity(Log, formatNormal, formatVerbose)
	for _, documenterLogger := range documenterLoggers {
		toggleVerbosity(documenterLogger, formatModuleNormal, formatModuleVerbose)
	}
}
