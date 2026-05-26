package lifx

import (
	"encoding/binary"
	"fmt"
	"gitlab.com/wwsean08/lifx-streamdeck/models"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.com/wwsean08/golifx"
)

const (
	sceneCommandAttempts = 3
	sceneRetryDelay      = 75 * time.Millisecond
)

// Client represents our client data

type lifxClient struct {
	deviceMap       map[string]*golifx.Device
	gSettings       *models.GlobalSettings
	writeLock       *sync.Mutex
	sceneLock       *sync.Mutex
	sceneGeneration uint64
	sceneCancel     chan struct{}
	logger          func(string)
	broadcastOnly   map[string]bool
}

func NewLifxController() Controller {
	client := new(lifxClient)
	client.deviceMap = make(map[string]*golifx.Device)
	client.writeLock = new(sync.Mutex)
	client.sceneLock = new(sync.Mutex)
	client.broadcastOnly = make(map[string]bool)
	return client
}

type Controller interface {
	DiscoverDevices() error
	GetCurrentColor(*golifx.Device) (*golifx.HSBK, error)
	GetCurrentState(*golifx.Device) (*golifx.DeviceState, error)
	GetPowerState(*golifx.Device) (bool, error)
	SetPowerState(*models.PowerSettings, bool) error
	TogglePowerState(*models.ToggleSettings) error
	SetColor(*models.ColorSettings) error
	SetBrightness(*models.BrightnessSettings) error
	SetWaveform(*models.WaveFormSettings) error
	SetScene(*models.SceneSettings) error
	UpdateGlobalSettings(*models.GlobalSettings) error
	GetDevices() map[string]*golifx.Device
	Debug() (*models.Debug, error)
}

func (c *lifxClient) SetLogger(logger func(string)) {
	c.logger = logger
}

func (c *lifxClient) logf(format string, args ...interface{}) {
	if c.logger != nil {
		c.logger(fmt.Sprintf(format, args...))
	}
}

func (c *lifxClient) DiscoverDevices() error {
	devices, err := golifx.LookupDevices()
	if err != nil {
		c.logf("lifx discover broadcast failed: %v", err)
		return err
	}
	c.clearDeviceMap()
	c.addDeviceToMap(devices)
	c.logf("lifx discover broadcast found=%d", len(devices))
	for _, device := range devices {
		c.logf("lifx discover device mac=%s ip=%v", device.MacAddress(), device.IP())
	}
	if c.gSettings != nil && c.gSettings.DirectComm {
		go c.refreshDirectReachability(devices)
	}

	if c.gSettings == nil {
		return nil
	}
	var wg sync.WaitGroup
	for _, ip := range c.gSettings.CustomDevices {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			q := net.ParseIP(ip)
			addr := &net.IPAddr{q, ""}
			device, err := golifx.LookupDeviceByIP(addr)
			if err != nil {
				c.logf("lifx custom device lookup failed ip=%s err=%v", ip, err)
				return
			}
			c.logf("lifx custom device lookup ok ip=%s mac=%s", ip, device.MacAddress())
			c.addDeviceToMap([]*golifx.Device{device})
		}(ip)
	}
	wg.Wait()

	return nil
}

func (c *lifxClient) refreshDirectReachability(devices []*golifx.Device) {
	var wg sync.WaitGroup
	for _, device := range devices {
		if device == nil || device.IP() == nil {
			continue
		}
		wg.Add(1)
		go func(device *golifx.Device) {
			defer wg.Done()
			if _, err := golifx.LookupDeviceByIP(device.IP()); err != nil {
				c.markBroadcastOnly(device.MacAddress())
				c.logf("lifx direct probe failed mac=%s ip=%v err=%v", device.MacAddress(), device.IP(), err)
				return
			}
			c.clearBroadcastOnly(device.MacAddress())
			c.logf("lifx direct probe ok mac=%s ip=%v", device.MacAddress(), device.IP())
		}(device)
	}
	wg.Wait()
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

func (c *lifxClient) getDevice(mac string) *golifx.Device {
	if c.writeLock == nil {
		c.writeLock = new(sync.Mutex)
	}
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	return c.deviceMap[mac]
}

func (c *lifxClient) markBroadcastOnly(mac string) {
	if c.writeLock == nil {
		c.writeLock = new(sync.Mutex)
	}
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	if c.broadcastOnly == nil {
		c.broadcastOnly = make(map[string]bool)
	}
	c.broadcastOnly[mac] = true
}

func (c *lifxClient) clearBroadcastOnly(mac string) {
	if c.writeLock == nil {
		c.writeLock = new(sync.Mutex)
	}
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	if c.broadcastOnly == nil {
		return
	}
	delete(c.broadcastOnly, mac)
}

func (c *lifxClient) shouldUseBroadcastOnly(mac string) bool {
	if c.writeLock == nil {
		c.writeLock = new(sync.Mutex)
	}
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	return c.broadcastOnly != nil && c.broadcastOnly[mac]
}

func (c *lifxClient) GetCurrentColor(device *golifx.Device) (*golifx.HSBK, error) {
	cs, err := c.GetCurrentState(device)
	if err != nil {
		return nil, err
	}
	return cs.Color, nil
}

func (c *lifxClient) GetCurrentState(device *golifx.Device) (*golifx.DeviceState, error) {
	if device == nil {
		return nil, fmt.Errorf("cannot get color for missing device")
	}
	if c.shouldUseBroadcastOnly(device.MacAddress()) {
		c.logf("lifx get color using broadcast-only mac=%s", device.MacAddress())
		return getCurrentStateViaBroadcast(device.MacAddress())
	}
	cs, err := device.GetColorState()
	if err != nil {
		mac := device.MacAddress()
		c.logf("lifx get color direct failed mac=%s err=%v", mac, err)
		fallback, fallbackErr := deviceFromMac(mac)
		if fallbackErr != nil {
			c.logf("lifx get color broadcast fallback unavailable mac=%s err=%v", mac, fallbackErr)
			return nil, err
		}
		cs, fallbackErr = fallback.GetColorState()
		if fallbackErr != nil {
			c.logf("lifx get color broadcast failed mac=%s err=%v", mac, fallbackErr)
			return nil, err
		}
		c.markBroadcastOnly(mac)
		c.logf("lifx get color broadcast ok mac=%s", mac)
	}
	if cs == nil || cs.Color == nil {
		return nil, fmt.Errorf("device returned no color state")
	}
	return cs, nil
}

func (c *lifxClient) GetPowerState(device *golifx.Device) (bool, error) {
	if device == nil {
		return false, fmt.Errorf("cannot get power for missing device")
	}
	if c.shouldUseBroadcastOnly(device.MacAddress()) {
		c.logf("lifx get power using broadcast-only mac=%s", device.MacAddress())
		return getPowerStateViaBroadcast(device.MacAddress())
	}
	state, err := device.GetPowerState()
	if err != nil {
		mac := device.MacAddress()
		c.logf("lifx get power direct failed mac=%s err=%v", mac, err)
		state, fallbackErr := getPowerStateViaBroadcast(mac)
		if fallbackErr != nil {
			c.logf("lifx get power broadcast failed mac=%s err=%v", mac, fallbackErr)
			return false, err
		}
		c.markBroadcastOnly(mac)
		c.logf("lifx get power broadcast ok mac=%s", mac)
		return state, nil
	}
	return state, nil
}

func (c *lifxClient) SetPowerState(settings *models.PowerSettings, on bool) error {
	c.cancelScene()
	devices, err := c.resolveDevices(settings.Devices)
	if err != nil {
		return err
	}
	c.logf("lifx power start on=%t requested=%d valid=%d transition=%d", on, len(settings.Devices), len(devices), settings.Transition)
	for _, lifxDevice := range devices {
		go func(lifxDevice *golifx.Device) {
			mac := lifxDevice.MacAddress()
			if c.shouldUseBroadcastOnly(mac) {
				if err := setPowerViaBroadcast(mac, on, settings.Transition); err != nil {
					c.logf("lifx power broadcast-only failed mac=%s on=%t err=%v", mac, on, err)
					return
				}
				c.logf("lifx power broadcast-only ok mac=%s on=%t", mac, on)
				return
			}
			if err := lifxDevice.SetPowerDurationState(on, settings.Transition); err != nil {
				c.logf("lifx power direct failed mac=%s on=%t err=%v", mac, on, err)
				if fallbackErr := setPowerViaBroadcast(mac, on, settings.Transition); fallbackErr != nil {
					c.logf("lifx power broadcast failed mac=%s on=%t err=%v", mac, on, fallbackErr)
					return
				}
				c.markBroadcastOnly(mac)
				c.logf("lifx power broadcast ok mac=%s on=%t", mac, on)
				return
			}
			c.logf("lifx power direct ok mac=%s on=%t", mac, on)
		}(lifxDevice)
	}
	return nil
}

func (c *lifxClient) TogglePowerState(settings *models.ToggleSettings) error {
	c.cancelScene()
	devices, err := c.resolveDevices(settings.Devices)
	if err != nil {
		return err
	}
	for _, lifxDevice := range devices {
		go func(lifxDevice *golifx.Device) {
			state, err := lifxDevice.GetPowerState()
			if err != nil {
				return
			}
			_ = lifxDevice.SetPowerDurationState(!state, settings.Transition)
		}(lifxDevice)
	}
	return nil
}

func (c *lifxClient) SetColor(settings *models.ColorSettings) error {
	c.cancelScene()
	hue, saturation, brightness, kelvin, transition := settings.Color.GenerateLIFXValues()

	hsbk := golifx.HSBK{
		Hue:        hue,
		Saturation: saturation,
		Brightness: brightness,
		Kelvin:     kelvin,
	}

	devices, err := c.resolveDevices(settings.Devices)
	if err != nil {
		return err
	}
	for _, lifxDevice := range devices {
		go func(lifxDevice *golifx.Device, hsbk golifx.HSBK) {
			_ = lifxDevice.SetColorState(&hsbk, transition)
		}(lifxDevice, hsbk)
	}

	return nil
}

func (c *lifxClient) SetBrightness(settings *models.BrightnessSettings) error {
	c.cancelScene()
	devices, err := c.resolveDevices(settings.Devices)
	if err != nil {
		return err
	}
	for _, lifxDevice := range devices {
		go func(lifxDevice *golifx.Device) {
			current, err := lifxDevice.GetColorState()
			if err != nil {
				return
			}
			hsbk := current.Color
			hsbk.Brightness = uint16(settings.Brightness * 655)
			_ = lifxDevice.SetColorState(hsbk, settings.Transition)
		}(lifxDevice)
	}
	return nil
}

func (c *lifxClient) SetWaveform(settings *models.WaveFormSettings) error {
	c.cancelScene()
	hue, saturation, brightness, kelvin, _ := settings.AlternateColor.GenerateLIFXValues()
	hsbk := &golifx.HSBK{
		Hue:        hue,
		Saturation: saturation,
		Brightness: brightness,
		Kelvin:     kelvin,
	}
	devices, err := c.resolveDevices(settings.Devices)
	if err != nil {
		return err
	}
	for _, lifxDevice := range devices {
		go func(lifxDevice *golifx.Device) {
			_, _ = lifxDevice.SetWaveform(settings.IsTransient, hsbk, settings.Period,
				settings.Cycles, settings.GetSkewRatio(), settings.Waveform)
		}(lifxDevice)
	}
	return nil
}

func (c *lifxClient) SetScene(settings *models.SceneSettings) error {
	if settings == nil || len(settings.Devices) == 0 {
		return nil
	}
	normalizeSceneSettings(settings)
	cancel := c.startScene()
	commands := buildSceneCommands(settings)
	devices, err := c.resolveDevices(settings.Devices)
	if err != nil {
		c.cancelScene()
		return err
	}
	c.logf("lifx scene start preset=%s reverse=%v devices=%d commands=%d", settings.Preset, settings.Reverse, len(devices), len(commands))

	go c.runScene(cancel, devices, commands)
	return nil
}

func (c *lifxClient) resolveDevices(settingsDevices []models.Device) (map[string]*golifx.Device, error) {
	devices := make(map[string]*golifx.Device)
	missing := false
	for _, device := range settingsDevices {
		if _, err := macToUint64(device.Mac); err != nil {
			c.logf("lifx resolve skip invalid mac=%q name=%q", device.Mac, device.Name)
			continue
		}
		if c.getDevice(device.Mac) == nil {
			missing = true
			break
		}
	}
	if missing {
		_ = c.DiscoverDevices()
	}

	for _, device := range settingsDevices {
		if _, err := macToUint64(device.Mac); err != nil {
			continue
		}
		lifxDevice := c.getDevice(device.Mac)
		if lifxDevice == nil {
			tmpDevice, err := deviceFromMac(device.Mac)
			if err != nil {
				continue
			}
			c.logf("lifx resolve fallback mac=%s using broadcast", device.Mac)
			c.addDeviceToMap([]*golifx.Device{tmpDevice})
			lifxDevice = tmpDevice
		}
		devices[device.Mac] = lifxDevice
	}
	if len(settingsDevices) > 0 && len(devices) == 0 {
		return nil, fmt.Errorf("no valid LIFX devices selected")
	}
	return devices, nil
}

func (c *lifxClient) startScene() chan struct{} {
	if c.sceneLock == nil {
		c.sceneLock = new(sync.Mutex)
	}
	c.sceneLock.Lock()
	defer c.sceneLock.Unlock()
	if c.sceneCancel != nil {
		close(c.sceneCancel)
	}
	c.sceneGeneration++
	c.sceneCancel = make(chan struct{})
	return c.sceneCancel
}

func (c *lifxClient) cancelScene() {
	if c.sceneLock == nil {
		c.sceneLock = new(sync.Mutex)
	}
	c.sceneLock.Lock()
	defer c.sceneLock.Unlock()
	if c.sceneCancel != nil {
		close(c.sceneCancel)
		c.sceneCancel = nil
	}
	c.sceneGeneration++
}

func (c *lifxClient) runScene(cancel <-chan struct{}, devices map[string]*golifx.Device, commands []sceneCommand) {
	start := time.Now()
	workers := c.startSceneWorkers(cancel, devices)
	defer c.closeSceneWorkers(workers)

	for _, batch := range groupSceneCommandsByDelay(commands) {
		if len(batch) == 0 {
			continue
		}
		wait := time.Until(start.Add(batch[0].Delay))
		if wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-timer.C:
			case <-cancel:
				timer.Stop()
				return
			}
		}
		select {
		case <-cancel:
			return
		default:
		}
		c.logf("lifx scene batch delay=%s commands=%d", batch[0].Delay, len(batch))
		for _, command := range batch {
			worker := workers[command.DeviceMac]
			if worker == nil {
				continue
			}
			select {
			case worker.commands <- command:
			case <-cancel:
				return
			}
		}
	}
}

type sceneWorker struct {
	commands chan sceneCommand
	done     chan struct{}
}

func (c *lifxClient) startSceneWorkers(cancel <-chan struct{}, devices map[string]*golifx.Device) map[string]*sceneWorker {
	workers := make(map[string]*sceneWorker, len(devices))
	for mac, device := range devices {
		if device == nil {
			continue
		}
		worker := &sceneWorker{
			commands: make(chan sceneCommand, 8),
			done:     make(chan struct{}),
		}
		workers[mac] = worker
		go func(device *golifx.Device, worker *sceneWorker) {
			defer close(worker.done)
			executeDeviceSceneCommands(cancel, device, worker.commands, c.logf, c.shouldUseBroadcastOnly, c.markBroadcastOnly)
		}(device, worker)
	}
	return workers
}

func (c *lifxClient) closeSceneWorkers(workers map[string]*sceneWorker) {
	for _, worker := range workers {
		close(worker.commands)
	}
	for _, worker := range workers {
		<-worker.done
	}
	c.logf("lifx scene finished")
}

func groupSceneCommandsByDelay(commands []sceneCommand) [][]sceneCommand {
	if len(commands) == 0 {
		return nil
	}
	groups := make([][]sceneCommand, 0)
	currentDelay := commands[0].Delay
	current := make([]sceneCommand, 0)
	for _, command := range commands {
		if command.Delay != currentDelay {
			groups = append(groups, current)
			currentDelay = command.Delay
			current = make([]sceneCommand, 0)
		}
		current = append(current, command)
	}
	return append(groups, current)
}

func groupSceneCommandsByDevice(commands []sceneCommand) map[string][]sceneCommand {
	byDevice := make(map[string][]sceneCommand)
	for _, command := range commands {
		byDevice[command.DeviceMac] = append(byDevice[command.DeviceMac], command)
	}
	return byDevice
}

func executeDeviceSceneCommands(cancel <-chan struct{}, device *golifx.Device, commands <-chan sceneCommand, logf func(string, ...interface{}), shouldUseBroadcastOnly func(string) bool, markBroadcastOnly func(string)) {
	for command := range commands {
		select {
		case <-cancel:
			return
		default:
		}
		if err := sendSceneCommand(cancel, device, command, logf, shouldUseBroadcastOnly, markBroadcastOnly); err != nil {
			logf("lifx scene command failed mac=%s delay=%s color=%t power=%t err=%v", device.MacAddress(), command.Delay, command.Color != nil, command.Power != nil, err)
			continue
		}
		logf("lifx scene command ok mac=%s delay=%s color=%t power=%t", device.MacAddress(), command.Delay, command.Color != nil, command.Power != nil)
	}
}

func sendSceneCommand(cancel <-chan struct{}, device *golifx.Device, command sceneCommand, logf func(string, ...interface{}), shouldUseBroadcastOnly func(string) bool, markBroadcastOnly func(string)) error {
	mac := device.MacAddress()
	if shouldUseBroadcastOnly != nil && shouldUseBroadcastOnly(mac) {
		err := sendSceneCommandViaBroadcast(mac, command)
		if err == nil {
			logf("lifx scene broadcast-only ok mac=%s delay=%s color=%t power=%t", mac, command.Delay, command.Color != nil, command.Power != nil)
		}
		return err
	}

	var err error
	for attempt := 0; attempt < sceneCommandAttempts; attempt++ {
		err = sendSceneCommandToDevice(device, command)
		if err == nil {
			return nil
		}
		select {
		case <-cancel:
			return err
		case <-time.After(sceneRetryDelay):
		}
	}
	logf("lifx scene direct failed mac=%s delay=%s color=%t power=%t err=%v", device.MacAddress(), command.Delay, command.Color != nil, command.Power != nil, err)
	fallback, fallbackErr := deviceFromMac(device.MacAddress())
	if fallbackErr != nil {
		logf("lifx scene broadcast fallback unavailable mac=%s err=%v", device.MacAddress(), fallbackErr)
		return err
	}
	fallbackErr = sendSceneCommandToDevice(fallback, command)
	if fallbackErr == nil {
		if markBroadcastOnly != nil {
			markBroadcastOnly(device.MacAddress())
		}
		logf("lifx scene broadcast ok mac=%s delay=%s color=%t power=%t", device.MacAddress(), command.Delay, command.Color != nil, command.Power != nil)
		return nil
	}
	logf("lifx scene broadcast failed mac=%s delay=%s color=%t power=%t err=%v", device.MacAddress(), command.Delay, command.Color != nil, command.Power != nil, fallbackErr)
	return err
}

func sendSceneCommandViaBroadcast(mac string, command sceneCommand) error {
	fallback, err := deviceFromMac(mac)
	if err != nil {
		return err
	}
	return sendSceneCommandToDevice(fallback, command)
}

func sendSceneCommandToDevice(device *golifx.Device, command sceneCommand) error {
	if command.Power != nil {
		return device.SetPowerDurationState(*command.Power, command.Transition)
	}
	if command.Color != nil {
		return device.SetColorState(command.Color, command.Transition)
	}
	return nil
}

func setPowerViaBroadcast(mac string, on bool, transition uint32) error {
	fallback, err := deviceFromMac(mac)
	if err != nil {
		return err
	}
	return fallback.SetPowerDurationState(on, transition)
}

func getCurrentColorViaBroadcast(mac string) (*golifx.HSBK, error) {
	cs, err := getCurrentStateViaBroadcast(mac)
	if err != nil {
		return nil, err
	}
	return cs.Color, nil
}

func getCurrentStateViaBroadcast(mac string) (*golifx.DeviceState, error) {
	fallback, err := deviceFromMac(mac)
	if err != nil {
		return nil, err
	}
	cs, err := fallback.GetColorState()
	if err != nil {
		return nil, err
	}
	if cs == nil || cs.Color == nil {
		return nil, fmt.Errorf("device returned no color state")
	}
	return cs, nil
}

func getPowerStateViaBroadcast(mac string) (bool, error) {
	fallback, err := deviceFromMac(mac)
	if err != nil {
		return false, err
	}
	return fallback.GetPowerState()
}

func (c *lifxClient) UpdateGlobalSettings(settings *models.GlobalSettings) error {
	if settings == nil {
		settings = new(models.GlobalSettings)
	}
	c.gSettings = settings
	golifx.SetAlwaysBroadcast(!settings.DirectComm)
	golifx.SetCacheTTL(time.Minute)
	if strings.TrimSpace(settings.OutIP) != "" {
		golifx.SetOutboundIP(&settings.OutIP)
	} else {
		golifx.SetOutboundIP(nil)
	}
	return nil
}

func (c *lifxClient) GetDevices() map[string]*golifx.Device {
	if c.writeLock == nil {
		c.writeLock = new(sync.Mutex)
	}
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	devices := make(map[string]*golifx.Device, len(c.deviceMap))
	for mac, device := range c.deviceMap {
		devices[mac] = device
	}
	return devices
}

// macToUint64 takes a mac address string and converts it to a mac address that
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

// deviceFromMac takes a mac address and generates a LIFX Device from it.
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
