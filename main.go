//go:build !windows
// +build !windows

package main

import (
	"encoding/json"
	"gitlab.com/wwsean08/lifx-streamdeck/models"
	"gitlab.com/wwsean08/lifx-streamdeck/streamdeck"
	"log"
	"os"
	"syscall"
	"time"

	"gitlab.com/wwsean08/lifx-streamdeck/lifx"
)

var (
	version = "develop"
	commit  = "unknown"
)

func main() {
	f, err := openCrashLog(0600)
	if err == nil {
		redirectStderr(f)
	}
	args := os.Args[1:]
	port := args[1]
	uuid := args[3]
	info := args[7]
	appInfo := new(models.AppInfo)
	_ = json.Unmarshal([]byte(info), appInfo)
	appInfo.Version = version
	appInfo.Commit = commit
	lifxController := lifx.NewLifxController()
	_, err = streamdeck.NewClient(port, uuid, appInfo, lifxController)
	if err != nil {
		panic(err)
	}
	for {
		time.Sleep(time.Minute)
	}
}

// redirectStderr to the file passed in
func redirectStderr(f *os.File) {
	err := syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd()))
	if err != nil {
		log.Fatalf("Failed to redirect stderr to file: %v", err)
	}
}
