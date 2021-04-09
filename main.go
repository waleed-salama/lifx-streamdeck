// +build !windows

package main

import (
	"encoding/json"
	"gitlab.com/wwsean08/lifx-streamdeck/models"
	"gitlab.com/wwsean08/lifx-streamdeck/streamdeck"
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
	appInfo := new(models.AppInfo)
	json.Unmarshal([]byte(info), appInfo)
	lifxController := lifx.NewLifxController()
	_, err := streamdeck.NewClient(port, uuid, appInfo, lifxController)
	if err != nil {
		panic(err)
	}
	for {
		time.Sleep(time.Minute)
	}
}
