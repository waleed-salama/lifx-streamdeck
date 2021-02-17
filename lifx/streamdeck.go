package lifx

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"gitlab.com/wwsean08/golifx"
	"gitlab.com/wwsean08/streamdeck"
)

// OnPropertyInspectorDidAppear is called when the property inspector appears in order to send data to the property inspector
func (c *Client) OnPropertyInspectorDidAppear(msg streamdeck.PropertyInspectorDidAppearMsg) {
	c.sendDevicesToPropertyInspector(msg.Action, msg.Context, c.devices)
}

// OnWillAppear is called when an item is displayed on the stream deck
func (c *Client) OnWillAppear(msg streamdeck.WillAppearMsg) {
	if len(msg.Payload.Settings) == 0 {
		//this is brand new or unconfigured
		return
	}
	// MigrateActions settings if needed
	settings, err := c.MigrateActions(msg.Payload.Settings, msg.Action)
	if err != nil {
		c.sdClient.Log(err.Error())
		panic(err)
	}
	settingsMsg := streamdeck.SetSettingsMsg{
		Context: msg.Context,
		Event:   streamdeck.SetSettingsEvent,
		Payload: settings,
	}
	c.sdClient.SendMessage(settingsMsg)
}

// OnSendToPlugin is called when a message is sent to the plugin (generally from the property inspector)
func (c *Client) OnSendToPlugin(msg streamdeck.SendToPluginMsg) {
	switch msg.Payload["type"] {
	case "discovery":
		c.discoverDevices(msg.Action, msg.Context)
	case "getColor":
		c.getDevicesCurrentColor(msg.Action, msg.Context, msg.Payload["mac"].(string))
	default:
		c.sdClient.Log(fmt.Sprintf("Unknown message type recieved from Property Inspector, %s", msg.Payload["type"]))
	}
}

// OnKeyUp is called when the Stream Deck key is unpressed
func (c *Client) OnKeyUp(msg streamdeck.KeyUpMsg) {
	switch msg.Action {
	case ActionTurnOnDevice:
		c.turnLights(msg.Context, msg.Payload.Settings, true)
	case ActionTurnOffDevice:
		c.turnLights(msg.Context, msg.Payload.Settings, false)
	case ActionSetColor:
		c.setColor(msg.Context, msg.Payload.Settings)
	case ActionSetBrightness:
		c.setBrightness(msg.Context, msg.Payload.Settings)
	case ActionSetWaveform:
		c.setWaveform(msg.Context, msg.Payload.Settings)
	case ActionToggleDevice:
		c.togglePower(msg.Context, msg.Payload.Settings)
	case ActionDebug:
		c.generateDebug(msg.Context)
	default:
		c.SendWarnMessage(msg.Context)
		c.sdClient.Log(fmt.Sprintf("Unknown action received %s", msg.Action))
	}
}

func (c *Client) OnDidReceiveGlobalSettings(msg streamdeck.DidReceiveGlobalSettingsMsg) {
	settings := msg.Payload.Settings
	// generally this will be a no-op but makes logic simpler
	settings = c.MigrateGlobalSettings(settings)
	c.UpdateGlobalSettings(settings)
}

func (c *Client) sendOKMessage(context string) {
	msg := streamdeck.ShowOkMsg{
		Event:   streamdeck.ShowOkEvent,
		Context: context,
	}

	_ = c.sdClient.SendMessage(msg)
}

// SendWarnMessage sends a warning message to the Stream Deck showing a warning symbol on the key
func (c *Client) SendWarnMessage(context string) {
	msg := streamdeck.ShowAlertMsg{
		Event:   streamdeck.ShowAlertEvent,
		Context: context,
	}

	_ = c.sdClient.SendMessage(msg)
}

// DebugCallback is used during debugging in development, it won't run in production
func (c *Client) DebugCallback(msg []byte) {
	c.sdClient.Log(string(msg))
}

func (c *Client) sendColorStateToPropertyInspector(action, context string, hsbk *golifx.HSBK) {
	type colorState struct {
		Hue        int `json:"hue"`
		Saturation int `json:"saturation"`
		Brightness int `json:"brightness"`
		Kelvin     int `json:"kelvin"`
	}

	data := colorState{
		Hue:        int(hsbk.Hue / 182), // don't ask about the 182 :P
		Saturation: int(hsbk.Saturation / 655),
		Brightness: int(hsbk.Brightness / 655),
		Kelvin:     int(hsbk.Kelvin),
	}

	msg := streamdeck.SendToPropertyInspectorMsg{
		Action:  action,
		Context: context,
		Event:   streamdeck.SendToPropertyInspectorEvent,
		Payload: map[string]interface{}{
			"type":  "getColor",
			"color": data,
		},
	}
	err := c.sdClient.SendMessage(msg)
	if err != nil {
		c.sdClient.Log(err.Error())
	}
}

func (c *Client) sendDevicesToPropertyInspector(action string, context string, devices map[string]*golifx.Device) {
	type deviceInfo struct {
		Label string `json:"label"`
		Mac   string `json:"mac"`
	}
	data := make([]deviceInfo, 0)
	for mac, device := range devices {
		// Label for the users, mac for our usage
		label, err := device.GetLabel()
		if err != nil {
			label = device.MacAddress()
		}
		data = append(data, deviceInfo{Label: label, Mac: mac})
	}

	msg := streamdeck.SendToPropertyInspectorMsg{
		Action:  action,
		Context: context,
		Event:   streamdeck.SendToPropertyInspectorEvent,
		Payload: map[string]interface{}{
			"type":    "discovery",
			"devices": data,
		},
	}
	err := c.sdClient.SendMessage(msg)
	if err != nil {
		c.sdClient.Log(err.Error())
	}
}

func (c Client) UpdateGlobalSettings(settings map[string]interface{}) {
	var gSettings = new(GlobalSettings)
	err := mapstructure.Decode(settings, gSettings)
	if err != nil {
		c.sdClient.Log(err.Error())
		return
	}

	*c.globalSettings = *gSettings
	if gSettings.OutIP == "" {
		golifx.SetOutboundIP(nil)
	} else {
		golifx.SetOutboundIP(&gSettings.OutIP)
	}
}

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

	c.sendOKMessage(context)
}

type network struct {
	IP         string          `json:"ip_used"`
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
	return network{
		IP:         h,
		Interfaces: ifInfo,
	}, nil
}
