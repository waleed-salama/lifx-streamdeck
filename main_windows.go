package main

import (
	"encoding/json"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"gitlab.com/wwsean08/lifx-streamdeck/lifx"
)

var (
	version = "develop"
	commit  = "unknown"
)

func main() {
	viper.Set("application.version", version)
	viper.Set("application.commit", commit)
	f, err := os.OpenFile("crash_log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	redirectStderr(f)
	args := os.Args[1:]
	port := args[1]
	uuid := args[3]
	info := args[7]
	appInfo := new(lifx.AppInfo)
	json.Unmarshal([]byte(info), appInfo)
	client, err := lifx.NewClient(port, uuid, appInfo)
	if err != nil {
		panic(err)
	}
	go client.Init()
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
