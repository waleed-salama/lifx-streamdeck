package lifx

import (
	"encoding/json"
	"github.com/mostlygeek/arp"
	"gitlab.com/wwsean08/golifx"
	"gitlab.com/wwsean08/lifx-streamdeck/models"
	"net"
	"strings"
)

func (c *lifxClient) Debug() (*models.Debug, error) {
	retVal := new(models.Debug)

	netInfo, err := getNetworkInfo()
	if err != nil {
		return retVal, err
	}

	netData, err := json.Marshal(netInfo)
	if err != nil {
		return retVal, err
	}
	deviceData, err := json.Marshal(c.deviceMap)
	if err != nil {
		return retVal, err
	}

	retVal.DeviceData = deviceData
	retVal.NetInfo = netData

	return retVal, nil
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

			// Not sure why I need this bit shift but via testing this is what I determined I needed, may be fragile
			macInt = macInt >> 16
			tmp := new(golifx.Device)
			tmp.SetHardwareAddress(macInt)
			_, err = tmp.GetLabel()
			if err == nil {
				lifxIPs[ip] = true
			}
		}
	}
	return lifxIPs
}
