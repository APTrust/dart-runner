package core

import (
	"database/sql"
	"log"
	"path/filepath"

	"github.com/APTrust/dart-runner/constants"
	"github.com/APTrust/dart-runner/util"
	"github.com/op/go-logging"
	_ "modernc.org/sqlite"
)

var Dart *DartContext

type DartContext struct {
	DB          *sql.DB
	Log         *logging.Logger
	Paths       *util.Paths
	RuntimeMode string
}

func init() {
	paths := util.NewPaths()
	Dart = &DartContext{
		DB:          initDB(paths),
		Paths:       paths,
		Log:         util.GetLogger(logging.DEBUG),
		RuntimeMode: constants.ModeDartRunner, // this should be set on startup
	}
	InitSchema()
}

func initDB(paths *util.Paths) *sql.DB {
	// Note: We're using pure go sqlite from modernc.org,
	// so the driver name is "sqlite". If we were using
	// github.com/mattn/go-sqlite3, the driver name
	// would have to change to "sqlite3". Both are
	// compatible with sqlite3.
	db, err := sql.Open("sqlite", DataFilePath())
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func LogFilePath() string {
	paths := util.NewPaths()
	return filepath.Join(paths.LogDir, "dart.log")
}

func DataFilePath() string {
	paths := util.NewPaths()
	dbPath := filepath.Join(paths.DataDir, "dart.db")
	// Run tests in an in-memory db, so we don't pollute
	// our actual dart db.
	if util.TestsAreRunning() {
		dbPath = ":memory:"
	}
	return dbPath
}
