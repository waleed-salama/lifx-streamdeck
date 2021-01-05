// +build !windows

package main

import (
	"encoding/json"
	"os"
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
