package util

import (
	"fmt"
	stdlog "log"
	"os"
	"path/filepath"

	"github.com/APTrust/dart-runner/constants"
	"github.com/op/go-logging"
)

var log *logging.Logger

func GetLogger(level logging.Level) *logging.Logger {
	if log != nil {
		return log
	}
	paths := NewPaths()

	if !FileExists(paths.LogDir) {
		err := os.MkdirAll(paths.LogDir, 0755)
		if err != nil && err != os.ErrExist {
			panic(err)
		}
	}

	logFile := filepath.Join(paths.LogDir, "dart.log")

	rotatedTo, rotationError := RotateCurrentLog(paths.LogDir, logFile, "dart", constants.MaxLogFileSize)

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

	if rotatedTo != "" {
		log.Infof("Last log was rotated to %s", rotatedTo)
	}
	if rotationError != nil {
		log.Warningf("Tried to rotate last log, but got error: %v", rotationError)
	}

	return log
}

// RotateCurrentLog rotates the DART log it's larger than the value of maxSize.
func RotateCurrentLog(pathToLogDir, pathToCurrentLog, logNamePrefix string, maxSize int64) (string, error) {
	fileInfo, err := os.Stat(pathToCurrentLog)
	if err != nil {
		return "", err
	}

	if fileInfo.Size() < maxSize {
		return "", nil
	}

	files, err := filepath.Glob(filepath.Join(pathToLogDir, fmt.Sprintf("%s_*.log", logNamePrefix)))
	if err != nil {
		return "", nil
	}

	nextFileName := fmt.Sprintf("%s_%04d.log", logNamePrefix, len(files)+1)
	pathToNextFile := filepath.Join(pathToLogDir, nextFileName)

	err = os.Rename(pathToCurrentLog, pathToNextFile)

	return pathToNextFile, err
}
