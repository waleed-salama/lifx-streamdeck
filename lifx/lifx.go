package lifx

import (
	"encoding/binary"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
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
	powerSettings := new(PowerSettings)
	err := mapstructure.Decode(settings, powerSettings)
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}
	for _, device := range powerSettings.Devices {
		if c.devices[device.Mac] == nil {
			tmpDevice := golifx.Device{}
			address, err := macToUint64(device.Mac)
			if err != nil {
				c.sdClient.Log(err.Error())
				c.SendWarnMessage(context)
			}
			// Not sure why I need this bit shift but via testing this is what I determined I needed, may be fragile
			address = address >> 16
			tmpDevice.SetHardwareAddress(address)
			c.devices[device.Mac] = &tmpDevice
		}
		device := c.devices[device.Mac]
		_ = device.SetPowerState(on)
	}
}

func (c *Client) setColor(context string, settings map[string]interface{}) {
	colorSettings := new(ColorSettings)
	err := mapstructure.Decode(settings, colorSettings)
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}
	hue, saturation, brightness, kelvin, transition := colorSettings.Color.GenerateLIFXValues()

	hsbk := golifx.HSBK{
		Hue:        hue,
		Saturation: saturation,
		Brightness: brightness,
		Kelvin:     kelvin,
	}

	for _, device := range colorSettings.Devices {
		if c.devices[device.Mac] == nil {
			tmpDevice := golifx.Device{}
			mac, _ := macToUint64(device.Mac)
			tmpDevice.SetHardwareAddress(mac)
			c.devices[device.Mac] = &tmpDevice
		}
		device := c.devices[device.Mac]
		_ = device.SetColorState(&hsbk, transition)

	}
}

func (c *Client) setBrightness(context string, settings map[string]interface{}) {
	brightnessSettings := new(BrightnessSettings)
	err := mapstructure.Decode(settings, brightnessSettings)
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}

	for _, device := range brightnessSettings.Devices {
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
			hsbk.Brightness = uint16(brightnessSettings.Brightness * 655)
			_ = device.SetColorState(hsbk, brightnessSettings.Transition)
		}(device.Mac)
	}
}

func (c *Client) setWaveform(context string, settings map[string]interface{}) {
	waveFormSettings := new(WaveFormSettings)
	err := mapstructure.Decode(settings, waveFormSettings)
	if err != nil {
		c.sdClient.Log(err.Error())
		c.SendWarnMessage(context)
		return
	}

}

//macToUint64 takes a mac address string and converts it to a mac address that
// LIFX will understand
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
