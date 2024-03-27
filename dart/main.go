package main

import (
	"flag"

	"github.com/APTrust/dart-runner/server"
)

// Version value is injected at build time.
var Version string

func main() {
	port := flag.Int("Port", 8080, "Which port should DART listen on?")
	flag.Parse()
	server.SetVersion(Version)
	server.Run(*port)
}
