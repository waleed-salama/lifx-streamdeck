package lifx

import (
	"encoding/binary"
	"gitlab.com/wwsean08/lifx-streamdeck/models"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.com/wwsean08/golifx"
)

// Client represents our client data

type lifxClient struct {
	deviceMap map[string]*golifx.Device
	gSettings *models.GlobalSettings
	writeLock *sync.Mutex
}

func NewLifxController() Controller {
	client := new(lifxClient)
	client.deviceMap = make(map[string]*golifx.Device)
	client.writeLock = new(sync.Mutex)
	return client
}

type Controller interface {
	DiscoverDevices() error
	GetCurrentColor(*golifx.Device) (*golifx.HSBK, error)
	SetPowerState(*models.PowerSettings, bool) error
	TogglePowerState(*models.ToggleSettings) error
	SetColor(*models.ColorSettings) error
	SetBrightness(*models.BrightnessSettings) error
	SetWaveform(*models.WaveFormSettings) error
	UpdateGlobalSettings(*models.GlobalSettings) error
	GetDevices() map[string]*golifx.Device
	Debug() (*models.Debug, error)
}

func (c *lifxClient) DiscoverDevices() error {
	devices, err := golifx.LookupDevices()
	if err != nil {
		return err
	}
	c.clearDeviceMap()
	c.addDeviceToMap(devices)

	for _, ip := range c.gSettings.CustomDevices {
		go func(ip string) {
			q := net.ParseIP(ip)
			addr := &net.IPAddr{q, ""}
			device, err := golifx.LookupDeviceByIP(addr)
			if err != nil {
				return
			}
			c.addDeviceToMap([]*golifx.Device{device})
		}(ip)
	}

	return nil
}

func (c *lifxClient) clearDeviceMap() {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	c.deviceMap = make(map[string]*golifx.Device)
}

func (c *lifxClient) addDeviceToMap(devices []*golifx.Device) {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	if c.deviceMap == nil {
		c.deviceMap = make(map[string]*golifx.Device)
	}
	for _, device := range devices {
		c.deviceMap[device.MacAddress()] = device
	}
}

func (c *lifxClient) GetCurrentColor(device *golifx.Device) (*golifx.HSBK, error) {
	cs, _ := device.GetColorState()
	return cs.Color, nil
}

func (c *lifxClient) SetPowerState(settings *models.PowerSettings, on bool) error {
	for _, device := range settings.Devices {
		if c.deviceMap[device.Mac] == nil {
			tmpDevice, err := deviceFromMac(device.Mac)
			if err != nil {
				return err
			}
			c.addDeviceToMap([]*golifx.Device{tmpDevice})
		}
		lifxDevice := c.deviceMap[device.Mac]
		go func(powerSettings *models.PowerSettings) {
			_ = lifxDevice.SetPowerDurationState(on, settings.Transition)
		}(settings)
	}
	return nil
}

func (c *lifxClient) TogglePowerState(settings *models.ToggleSettings) error {
	for _, device := range settings.Devices {
		if c.deviceMap[device.Mac] == nil {
			tmpDevice, err := deviceFromMac(device.Mac)
			if err != nil {
				return err
			}
			c.addDeviceToMap([]*golifx.Device{tmpDevice})
		}
		lifxDevice := c.deviceMap[device.Mac]
		go func() {
			state, err := lifxDevice.GetPowerState()
			if err != nil {
				return
			}
			_ = lifxDevice.SetPowerDurationState(!state, settings.Transition)
		}()
	}
	return nil
}

func (c *lifxClient) SetColor(settings *models.ColorSettings) error {
	hue, saturation, brightness, kelvin, transition := settings.Color.GenerateLIFXValues()

	hsbk := golifx.HSBK{
		Hue:        hue,
		Saturation: saturation,
		Brightness: brightness,
		Kelvin:     kelvin,
	}

	for _, device := range settings.Devices {
		if c.deviceMap[device.Mac] == nil {
			tmpDevice, err := deviceFromMac(device.Mac)
			if err != nil {
				return err
			}
			c.addDeviceToMap([]*golifx.Device{tmpDevice})
		}
		lifxDevice := c.deviceMap[device.Mac]
		go func(settings *models.ColorSettings, hsbk golifx.HSBK) {
			_ = lifxDevice.SetColorState(&hsbk, transition)
		}(settings, hsbk)
	}

	return nil
}

func (c *lifxClient) SetBrightness(settings *models.BrightnessSettings) error {
	for _, device := range settings.Devices {
		if c.deviceMap[device.Mac] == nil {
			tmpDevice, err := deviceFromMac(device.Mac)
			if err != nil {
				return err
			}
			c.addDeviceToMap([]*golifx.Device{tmpDevice})
		}
		lifxDevice := c.deviceMap[device.Mac]

		go func(settings *models.BrightnessSettings) {
			current, err := lifxDevice.GetColorState()
			if err != nil {
				return
			}
			hsbk := current.Color
			hsbk.Brightness = uint16(settings.Brightness * 655)
			_ = lifxDevice.SetColorState(hsbk, settings.Transition)
		}(settings)
	}
	return nil
}

func (c *lifxClient) SetWaveform(settings *models.WaveFormSettings) error {
	hue, saturation, brightness, kelvin, _ := settings.AlternateColor.GenerateLIFXValues()
	hsbk := &golifx.HSBK{
		Hue:        hue,
		Saturation: saturation,
		Brightness: brightness,
		Kelvin:     kelvin,
	}
	for _, device := range settings.Devices {
		if c.deviceMap[device.Mac] == nil {
			tmpDevice, err := deviceFromMac(device.Mac)
			if err != nil {
				return err
			}
			c.addDeviceToMap([]*golifx.Device{tmpDevice})
		}
		lifxDevice := c.deviceMap[device.Mac]
		go func(settings *models.WaveFormSettings) {
			_, _ = lifxDevice.SetWaveform(settings.IsTransient, hsbk, settings.Period,
				settings.Cycles, settings.GetSkewRatio(), settings.Waveform)
		}(settings)
	}
	return nil
}

func (c *lifxClient) UpdateGlobalSettings(settings *models.GlobalSettings) error {
	c.gSettings = settings
	golifx.SetAlwaysBroadcast(!settings.DirectComm)
	golifx.SetCacheTTL(time.Minute)
	if strings.TrimSpace(settings.OutIP) != "" {
		golifx.SetOutboundIP(&settings.OutIP)
	}
	return nil
}

func (c *lifxClient) GetDevices() map[string]*golifx.Device {
	return c.deviceMap
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

//deviceFromMac takes a mac address and generates a LIFX Device from it.
func deviceFromMac(mac string) (*golifx.Device, error) {
	tmpDevice := new(golifx.Device)
	address, err := macToUint64(mac)
	if err != nil {
		return nil, err
	}
	// Not sure why I need this bit shift but via testing this is what I determined I needed, may be fragile
	address = address >> 16
	tmpDevice.SetHardwareAddress(address)
	return tmpDevice, nil
}
