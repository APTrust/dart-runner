package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"time"

	"github.com/APTrust/dart-runner/server"
	"github.com/getlantern/systray"
)

// Version value is injected at build time.
var Version string

func main() {
	systray.Run(onReady, onExit)
}

func onExit() {
	// cleanup
}

func onReady() {
	port := flag.Int("port", 8444, "Which port should DART listen on?")
	version := flag.Bool("version", false, "Show version and exit.")
	flag.Parse()
	server.SetVersion(Version)
	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}

	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// This is MacOS only
	iconPath := path.Join(exePath, "..", "..", "Resources", "icon.icns")
	fmt.Println(iconPath)
	icon := getIcon(iconPath)
	fmt.Println(len(icon))
	systray.SetIcon(icon)
	systray.SetTitle("DART3")
	mQuit := systray.AddMenuItem("Quit", "Quit DART3")

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

	go server.Run(*port, true)
	time.Sleep(1000 * time.Millisecond)
	openBrowser(fmt.Sprintf("http://localhost:%d", *port))
}

func openBrowser(url string) {
	var err error

	// TODO: See if we can get command PID
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		// log.Fatal(err)
	}
}

// TODO: embed this instead of loading it?
func getIcon(s string) []byte {
	b, err := os.ReadFile(s)
	if err != nil {
		fmt.Print(err)
	}
	return b
}
