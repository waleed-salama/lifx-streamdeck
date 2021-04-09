package streamdeck

import (
	"archive/zip"
	"encoding/json"
	"github.com/mitchellh/go-homedir"
	"gitlab.com/wwsean08/lifx-streamdeck/models"
	"io/ioutil"
	"os"
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

	zipFile, err := os.OpenFile(fName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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

	c.sendOKMessage(context)
}
