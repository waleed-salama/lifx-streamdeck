package lifx

import (
	"encoding/binary"
	"fmt"
	"gitlab.com/wwsean08/golifx"
	"gitlab.com/wwsean08/streamdeck"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	sdClient *streamdeck.Client
	devices  map[string]*golifx.Device
}

func NewClient(port, uuid string) (*Client, error) {
	sdClient, err := streamdeck.NewClient(port, uuid)
	if err != nil {
		return nil, err
	}
	err = sdClient.Init()
	if err != nil {
		return nil, err
	}
	client := &Client{
		sdClient: sdClient,
	}
	return client, nil
}

func (c *Client) Init() {
	c.sdClient.SetOnKeyUpCallback(c.OnKeyUp)
	c.sdClient.SetSendToPluginCallback(c.OnSendToPlugin)
	c.sdClient.SetPropertyInspectorDidAppearCallback(c.OnPropertyInspectorDidAppear)
	c.sdClient.SetRawCallback(func(msg []byte) {
		c.sdClient.Log(string(msg))
	})
	_, _ = c.getAllDevices()
}

func (c *Client) OnPropertyInspectorDidAppear(msg streamdeck.PropertyInspectorDidAppearMsg) {
	c.sendDevicesToPropertyInspector(msg.Action, msg.Context, c.devices)
}

func (c *Client) OnSendToPlugin(msg streamdeck.SendToPluginMsg) {
	switch msg.Payload["type"] {
	case "discovery":
		c.discoverDevices(msg.Action, msg.Context)
	case "getColor":
		c.getDevicesCurrentColor(msg.Action, msg.Context, msg.Payload["mac"].(string))
	}
}

func (c *Client) OnKeyUp(msg streamdeck.KeyUpMsg) {
	switch msg.Action {
	case ActionTurnOnLight:
		c.turnLights(msg.Payload.Settings, true)
	case ActionTurnOffLight:
		c.turnLights(msg.Payload.Settings, false)
	case ActionSetColor:
		c.setColor(msg.Context, msg.Payload.Settings)
	default:
		c.SendWarnMessage(msg.Context)
		c.sdClient.Log(fmt.Sprintf("Unknown action received %s", msg.Action))
	}
}

func (c *Client) SendWarnMessage(context string) {
	msg := streamdeck.ShowAlertMsg{
		Event:   streamdeck.ShowAlertEvent,
		Context: context,
	}

	_ = c.sdClient.SendMessage(msg)
}

func (c *Client) discoverDevices(action, context string) {
	devices, err := c.getAllDevices()
	if err != nil {
		c.SendWarnMessage(context)
		c.sdClient.Log(err.Error())
	}
	c.sendDevicesToPropertyInspector(action, context, devices)
}

func (c *Client) getDevicesCurrentColor(action, context, mac string) {
	device := c.devices[mac]
	if device == nil {
		device = new(golifx.Device)
		address, err := macToUint64(mac)
		if err != nil {
			c.sdClient.Log(err.Error())
		}
		device.SetHardwareAddress(address)
	}
	cs, _ := device.GetColorState()
	c.sendColorStateToPropertyInspector(action, context, cs.Color)
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
			c.sdClient.Log(err.Error())
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

func (c *Client) getAllDevices() (map[string]*golifx.Device, error) {
	devices, err := golifx.LookupDevices()
	if err != nil {
		return nil, err
	}
	deviceMap := map[string]*golifx.Device{}
	for _, device := range devices {
		deviceMap[device.MacAddress()] = device
	}
	c.devices = deviceMap
	return c.devices, nil
}

func (c *Client) turnLights(settings map[string]interface{}, on bool) {
	golifx.SetTTL(1 * time.Millisecond)
	defer golifx.SetTTL(time.Millisecond * 500)
	for key := range settings {
		if c.devices[key] == nil {
			device := golifx.Device{}
			address, err := macToUint64(key)
			c.sdClient.Log(fmt.Sprintf("%d", address))
			if err != nil {
				c.sdClient.Log(err.Error())
			}
			// Not sure why I need this bit shift but via testing this is what I determined I needed, may be fragile
			address = address >> 16
			device.SetHardwareAddress(address)
			c.sdClient.Log(device.MacAddress())
			err = device.SetPowerState(on)
			if err != nil {
				c.sdClient.Log(err.Error())
			}
		} else {
			device := c.devices[key]
			_ = device.SetPowerState(on)
		}
	}
}

func (c *Client) setColor(context string, settings map[string]interface{}) {
	devices := settings["devices"].(map[string]interface{})
	if len(devices) == 0 {
		// no devices to change so return quickly
		return
	}
	color := settings["color"].(map[string]interface{})
	hue, err := strconv.ParseUint(color["hue"].(string), 10, 16)
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}
	saturation, err := strconv.ParseUint(color["saturation"].(string), 10, 16)
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}
	brightness, err := strconv.ParseUint(color["brightness"].(string), 10, 16)
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}
	kelvin, err := strconv.ParseUint(color["kelvin"].(string), 10, 16)
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}
	hsbk := golifx.HSBK{
		Hue:        uint16(hue * 182),
		Saturation: uint16(saturation * 655),
		Brightness: uint16(brightness * 655),
		Kelvin:     uint16(kelvin),
	}

	golifx.SetTTL(time.Millisecond * 1)
	defer golifx.SetTTL(time.Millisecond * 500)
	for key := range devices {
		var device *golifx.Device
		if c.devices[key] != nil {
			device = c.devices[key]
		} else {
			device = &golifx.Device{}
			mac, _ := macToUint64(key)
			device.SetHardwareAddress(mac)
		}
		_ = device.SetColorState(&hsbk, 0)
	}
}

func macToUint64(mac string) (uint64, error) {
	macMinusColons := strings.Replace(mac, ":", "", -1)
	decMac, err := strconv.ParseUint(macMinusColons, 16, 64)
	if err != nil {
		return 0, err
	}
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, decMac)
	return binary.LittleEndian.Uint64(b), nil
}
