package lifx

import (
	"encoding/binary"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gitlab.com/wwsean08/golifx"
	"gitlab.com/wwsean08/streamdeck"
)

// Client represents our client data
type Client struct {
	sdClient *streamdeck.Client
	devices  map[string]*golifx.Device
}

// NewClient creates a client for speaking with the Stream Deck websocket
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

// Init handles initializing the client, and setting up the initial discovery of devices
func (c *Client) Init() {
	golifx.SetAlwaysBroadcast(true)
	c.sdClient.SetOnKeyUpCallback(c.OnKeyUp)
	c.sdClient.SetSendToPluginCallback(c.OnSendToPlugin)
	c.sdClient.SetPropertyInspectorDidAppearCallback(c.OnPropertyInspectorDidAppear)
	c.sdClient.SetWillAppearCallback(c.OnWillAppear)
	if viper.GetString("application.version") == "develop" {
		c.sdClient.SetRawCallback(c.DebugCallback)
	}
	_, _ = c.getAllDevices()
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
			c.SendWarnMessage(context)
		}
		device.SetHardwareAddress(address)
	}
	cs, _ := device.GetColorState()
	c.sendColorStateToPropertyInspector(action, context, cs.Color)
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

func (c *Client) turnLights(context string, settings map[string]interface{}, on bool) {
	golifx.SetTTL(1 * time.Millisecond)
	defer golifx.SetTTL(time.Millisecond * 500)
	for key := range settings {
		if c.devices[key] == nil {
			device := golifx.Device{}
			address, err := macToUint64(key)
			if err != nil {
				c.sdClient.Log(err.Error())
				c.SendWarnMessage(context)
			}
			// Not sure why I need this bit shift but via testing this is what I determined I needed, may be fragile
			address = address >> 16
			device.SetHardwareAddress(address)
			c.devices[key] = &device
		}
		device := c.devices[key]
		_ = device.SetPowerState(on)
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
	transition, err := strconv.ParseUint(color["transition"].(string), 10, 32)
	hsbk := golifx.HSBK{
		Hue:        uint16(hue * 182),
		Saturation: uint16(saturation * 655),
		Brightness: uint16(brightness * 655),
		Kelvin:     uint16(kelvin),
	}

	golifx.SetTTL(time.Millisecond * 1)
	defer golifx.SetTTL(time.Millisecond * 500)
	for key := range devices {
		if c.devices[key] == nil {
			device := golifx.Device{}
			mac, _ := macToUint64(key)
			device.SetHardwareAddress(mac)
			c.devices[key] = &device
		}
		device := c.devices[key]
		_ = device.SetColorState(&hsbk, uint32(transition))

	}
}

func (c *Client) setBrightness(context string, settings map[string]interface{}) {
	devices := settings["devices"].(map[string]interface{})
	if len(devices) == 0 {
		// no devices to change so return quickly
		return
	}
	brightness, err := strconv.ParseUint(settings["brightness"].(string), 10, 16)
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}
	transition, err := strconv.ParseUint(settings["transition"].(string), 10, 32)
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}

	for key := range devices {
		// Due to having to do some lookups before making the settings changes
		// i'm running these each in their own thread, that way they (to the eye)
		//  happen simultaneously.
		go func(key string) {
			if c.devices[key] == nil {
				device := golifx.Device{}
				mac, _ := macToUint64(key)
				device.SetHardwareAddress(mac)
				c.devices[key] = &device
			}
			device := c.devices[key]
			current, err := device.GetColorState()
			if err != nil {
				c.sdClient.Log(err.Error())
				c.SendWarnMessage(context)
				return
			}
			hsbk := current.Color
			hsbk.Brightness = uint16(brightness * 655)
			_ = device.SetColorState(hsbk, uint32(transition))
		}(key)
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
