package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/APTrust/dart-runner/server"
)

// Version value is injected at build time.
var Version string

func main() {
	port := flag.Int("port", 8444, "Which port should DART listen on?")
	version := flag.Bool("version", false, "Show version and exit.")
	flag.Parse()
	server.SetVersion(Version)
	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}
	server.Run(*port, true)
}
