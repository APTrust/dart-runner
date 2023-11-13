package util

import (
	stdlog "log"
	"os"
	"path"

	"github.com/op/go-logging"
)

var log *logging.Logger

func GetLogger(level logging.Level) *logging.Logger {
	if log != nil {
		return log
	}
	paths := NewPaths()
	logFile := path.Join(paths.LogDir, "dart.log")
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	log = logging.MustGetLogger("DART")
	logging.SetLevel(level, "DART")

	logBackend := logging.NewLogBackend(f, "", stdlog.LstdFlags|stdlog.LUTC)
	format := logging.MustStringFormatter("[%{level}] %{message}")

	logging.SetFormatter(format)
	logging.SetBackend(logBackend)

	log.Info("DART started")

	return log
}
