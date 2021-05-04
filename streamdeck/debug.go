package streamdeck

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"gitlab.com/wwsean08/lifx-streamdeck/models"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
)

func (c Client) generateDebug(context string, debug *models.Debug) {
	// get app environment info
	envData, err := json.Marshal(c.appInfo)
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}

	// get global settings as json
	gSettingsData, err := json.Marshal(c.globalSettings)
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}

	crashData, err := ioutil.ReadFile("crash_log")
	if err != nil {
		// If the file doesn't exist, that's not a big deal
		if !strings.HasSuffix(err.Error(), "no such file or directory") {
			c.sdClient.Log(err.Error())
			c.SendWarnMessage(context)
			return
		}
	}

	fName, err := homedir.Expand("~/lifx-controls-debug.zip")
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}

	zipFile, err := os.OpenFile(fName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}
	defer zipFile.Close()
	zipw := zip.NewWriter(zipFile)
	defer zipw.Close()

	netFile, err := zipw.Create("net.json")
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}
	netFile.Write(debug.NetInfo)

	envFile, err := zipw.Create("env.json")
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}
	envFile.Write(envData)

	crashFile, err := zipw.Create("crash_log")
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}
	crashFile.Write(crashData)

	gSettingsFile, err := zipw.Create("gSettings.json")
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}
	gSettingsFile.Write(gSettingsData)

	devicesFile, err := zipw.Create("devices.json")
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}
	devicesFile.Write(debug.DeviceData)

	logData, err := getLogs()
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}
	for fileName, fileContents := range logData {
		tmp, _ := zipw.Create(fileName)
		tmp.Write(fileContents)
	}

	c.sendOKMessage(context)
}

func getLogs() (map[string][]byte, error) {
	logLocation := ""
	logData := make(map[string][]byte)
	var err error
	if runtime.GOOS == "windows" {
		logLocation, err = homedir.Expand("~/AppData/Roaming/Elgato/StreamDeck/logs/")
		if err != nil {
			return logData, err
		}
	} else if runtime.GOOS == "darwin" {
		logLocation, err = homedir.Expand("~/Library/Logs/StreamDeck/")
		if err != nil {
			return logData, err
		}
	}
	logFiles, err := ioutil.ReadDir(logLocation)
	if err != nil {
		return logData, err
	}
	for _, logFile := range logFiles {
		// only grab logs from my plugin
		if strings.HasPrefix(logFile.Name(), "dev.sean.lifx") {
			logData[logFile.Name()], err = ioutil.ReadFile(fmt.Sprintf("%s/%s", logLocation, logFile.Name()))
			if err == nil {
				return logData, err
			}
		}
	}

	return logData, err
}
