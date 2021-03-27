package lifx

import (
	"archive/zip"
	"encoding/json"
	"github.com/mitchellh/go-homedir"
	"github.com/mostlygeek/arp"
	"gitlab.com/wwsean08/golifx"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

func (c Client) generateDebug(context string) {
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

	netInfo, err := getNetworkInfo()
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}
	netData, err := json.Marshal(netInfo)
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
	deviceData, err := json.Marshal(c.devices)

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
	netFile.Write(netData)

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
	devicesFile.Write(deviceData)

	c.sendOKMessage(context)
}

type network struct {
	IP         string          `json:"ip_used"`
	LIFXIps    map[string]bool `json:"lifx_ips"`
	Interfaces []interfaceInfo `json:"interfaces"`
}

type interfaceInfo struct {
	IFName string     `json:"if_name"`
	Addrs  []net.Addr `json:"addrs"`
}

// leveraging the fact that i know how golifx works to get the info i want
func getNetworkInfo() (network, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return network{}, err
	}
	h, _, err := net.SplitHostPort(conn.LocalAddr().String())
	if err != nil {
		return network{}, err
	}
	iFaces, err := net.Interfaces()
	ifInfo := make([]interfaceInfo, len(iFaces))
	if err != nil {
		return network{}, err
	}
	for i, iFace := range iFaces {
		addrs, _ := iFace.Addrs()
		tmp := interfaceInfo{
			IFName: iFace.Name,
			Addrs:  addrs,
		}
		ifInfo[i] = tmp
	}
	lifxIPs := getAllLifxIPs()
	return network{
		IP:         h,
		LIFXIps:    lifxIPs,
		Interfaces: ifInfo,
	}, nil
}

func getAllLifxIPs() map[string]bool {
	arpTable := arp.Table()
	lifxIPs := make(map[string]bool)
	for ip, mac := range arpTable {
		// LIFX has been allocated D0:73:D5
		if strings.HasPrefix(strings.ToUpper(mac), "D0:73:D5") {
			// check for connectivity, assume it won't connect
			lifxIPs[ip] = false
			macInt, err := macToUint64(mac)
			if err != nil {
				println(err.Error())
			}
			// Not sure why I need this bit shift but via testing this is what I determined I needed, may be fragile
			macInt = macInt >> 16
			tmp := new(golifx.Device)
			tmp.SetHardwareAddress(macInt)
			_, err = tmp.GetLabel()
			if err != nil {
				println(err.Error())
			} else {
				lifxIPs[ip] = true
			}
		}
	}
	return lifxIPs
}
