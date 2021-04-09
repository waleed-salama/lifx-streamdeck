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
	f, err := os.OpenFile("crash_log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	redirectStderr(f)
	args := os.Args[1:]
	port := args[1]
	uuid := args[3]
	info := args[7]
	appInfo := new(models.AppInfo)
	json.Unmarshal([]byte(info), appInfo)
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

var (
	kernel32         = syscall.MustLoadDLL("kernel32.dll")
	procSetStdHandle = kernel32.MustFindProc("SetStdHandle")
)

func setStdHandle(stdhandle int32, handle syscall.Handle) error {
	r0, _, e1 := syscall.Syscall(procSetStdHandle.Addr(), 2, uintptr(stdhandle), uintptr(handle), 0)
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		}
		return syscall.EINVAL
	}
	return nil
}

// redirectStderr to the file passed in
func redirectStderr(f *os.File) {
	err := setStdHandle(syscall.STD_ERROR_HANDLE, syscall.Handle(f.Fd()))
	if err != nil {
		log.Fatalf("Failed to redirect stderr to file: %v", err)
	}
	// SetStdHandle does not affect prior references to stderr
	os.Stderr = f
}
