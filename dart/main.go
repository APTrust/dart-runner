package main

import (
	"flag"
	"fmt"
	"log"
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
	//fmt.Println(iconPath)
	icon := getIcon(iconPath)
	//fmt.Println(len(icon))
	systray.SetIcon(icon)
	systray.SetTitle("DART3")

	go server.Run(*port, true)
	time.Sleep(1000 * time.Millisecond)

	url := fmt.Sprintf("http://localhost:%d", *port)
	command := openBrowser(url)

	mQuit := systray.AddMenuItem("Quit", "Quit DART3")
	mView := systray.AddMenuItem("View in Browser", "View DART3 in browser window")
	go func() {
		for {
			select {
			case <-mView.ClickedCh:
				// TODO: The back end should know (from pings) if the browser window
				// is still open. If it is, send a message from the back end to focus
				// on the window.
				//
				// https://developer.mozilla.org/en-US/docs/Web/API/Window/focus
				// https://stackoverflow.com/questions/3478654/is-there-a-browser-event-for-the-window-getting-focus
				//
				// Front end can tell back end if it has focus:
				//
				// https://stackoverflow.com/questions/7389328/detect-if-browser-tab-has-focus
				//
				// See also:
				//
				// window.onblur
				// document.hidden
				// document.visibilityState
				// https://developer.mozilla.org/en-US/docs/Web/API/Document/visibilitychange_event
				// https://developer.mozilla.org/en-US/docs/Web/API/Notification
				//
				// If tab is not open, open a new one.
				//
				// Note that command.Process.Exited() causes segfault on MacOS.
				// Here, we're just blindly opening a new tab.
				command = openBrowser(url)
			case <-mQuit.ClickedCh:
				// TODO: Back end should notify user if there are running jobs.
				systray.Quit()
				command.Process.Kill()

			}
		}
	}()

}

func openBrowser(url string) *exec.Cmd {
	var err error
	var command *exec.Cmd

	// TODO: See if we can get command PID
	switch runtime.GOOS {
	case "linux":
		command = exec.Command("xdg-open", url)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		command = exec.Command("open", url)
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

	err = command.Start()
	if err != nil {
		log.Fatal(err)
	}
	return command
}

// TODO: embed this instead of loading it?
func getIcon(s string) []byte {
	b, err := os.ReadFile(s)
	if err != nil {
		fmt.Print(err)
	}
	return b
}
