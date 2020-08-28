// +build !windows

package main

import (
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
	client, err := lifx.NewClient(port, uuid)
	if err != nil {
		panic(err)
	}
	go client.Init()
	for {
		time.Sleep(time.Minute)
	}
}
