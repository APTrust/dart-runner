package core

import (
	"database/sql"
	"log"
	"path"

	"github.com/APTrust/dart-runner/util"
	_ "modernc.org/sqlite"
)

var Dart *DartContext

type DartContext struct {
	DB    *sql.DB
	Log   *util.Logger
	Paths *util.Paths
}

func init() {
	paths := util.NewPaths()
	Dart = &DartContext{
		DB:    initDB(paths),
		Paths: paths,
		Log:   util.GetLogger(util.LevelDebug),
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
	return path.Join(paths.LogDir, "dart.log")
}

func DataFilePath() string {
	paths := util.NewPaths()
	dbPath := path.Join(paths.DataDir, "dart.db")
	// Run tests in an in-memory db, so we don't pollute
	// our actual dart db.
	if util.TestsAreRunning() {
		dbPath = ":memory:"
		//dbPath = path.Join(paths.HomeDir, "Desktop", "dart.db")
	}
	return dbPath
}
