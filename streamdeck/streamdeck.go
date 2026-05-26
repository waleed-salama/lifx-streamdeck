package streamdeck

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/reugn/go-quartz/quartz"
	"github.com/skratchdot/open-golang/open"
	"gitlab.com/wwsean08/golifx"
	"gitlab.com/wwsean08/lifx-streamdeck/lifx"
	"gitlab.com/wwsean08/lifx-streamdeck/models"
	"gitlab.com/wwsean08/streamdeck"
	"time"
)

type Client struct {
	sdClient       *streamdeck.Client
	appInfo        *models.AppInfo
	uuid           string
	globalSettings *models.GlobalSettings
	controller     lifx.Controller
	scheduler      quartz.Scheduler
}

// NewClient creates a client for speaking with the Stream Deck websocket
func NewClient(port, uuid string, info *models.AppInfo, controller lifx.Controller) (*Client, error) {
	sdClient, err := streamdeck.NewClient(port, uuid)
	if err != nil {
		return nil, err
	}
	err = sdClient.Init()
	if err != nil {
		return nil, err
	}
	gSettings := new(models.GlobalSettings)
	client := &Client{
		sdClient:       sdClient,
		appInfo:        info,
		uuid:           uuid,
		globalSettings: gSettings,
		controller:     controller,
	}
	if logger, ok := controller.(interface{ SetLogger(func(string)) }); ok {
		logger.SetLogger(func(msg string) {
			client.sdClient.Log(msg)
		})
	}
	go func() {
		_ = client.controller.DiscoverDevices()
	}()
	client.sdClient.SetOnKeyUpCallback(client.OnKeyUp)
	client.sdClient.SetSendToPluginCallback(client.OnSendToPlugin)
	client.sdClient.SetPropertyInspectorDidAppearCallback(client.OnPropertyInspectorDidAppear)
	client.sdClient.SetWillAppearCallback(client.OnWillAppear)
	client.sdClient.SetDidReceiveGlobalSettingsCallback(client.OnDidReceiveGlobalSettings)
	if info.Version == "develop" {
		client.sdClient.SetRawCallback(client.DebugCallback)
	}

	// Grab global settings for initialization
	msg := streamdeck.GetGlobalSettingsMsg{
		Context: uuid,
		Event:   streamdeck.GetGlobalSettingsEvent,
	}
	_ = client.sdClient.SendMessage(msg)
	return client, nil
}

// OnPropertyInspectorDidAppear is called when the property inspector appears in order to send data to the property inspector
func (c *Client) OnPropertyInspectorDidAppear(msg streamdeck.PropertyInspectorDidAppearMsg) {
	c.sendDevicesToPropertyInspector(msg.Action, msg.Context, c.controller.GetDevices())
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
	}
	settingsMsg := streamdeck.SetSettingsMsg{
		Context: msg.Context,
		Event:   streamdeck.SetSettingsEvent,
		Payload: settings,
	}
	_ = c.sdClient.SendMessage(settingsMsg)
}

// OnSendToPlugin is called when a message is sent to the plugin (generally from the property inspector)
func (c *Client) OnSendToPlugin(msg streamdeck.SendToPluginMsg) {
	switch msg.Payload["type"] {
	case "discovery":
		err := c.controller.DiscoverDevices()
		if err != nil {
			c.SendWarnMessage(msg.Context)
			c.sdClient.Log(err.Error())
		}
		c.sendDevicesToPropertyInspector(msg.Action, msg.Context, c.controller.GetDevices())
	case "getColor":
		mac, ok := msg.Payload["mac"].(string)
		if !ok || mac == "" {
			c.SendWarnMessage(msg.Context)
			err := fmt.Errorf("getColor requested without a device mac")
			c.sdClient.Log(err.Error())
			c.sendColorErrorToPropertyInspector(msg.Action, msg.Context, err)
			return
		}
		color, err := c.controller.GetCurrentColor(c.controller.GetDevices()[mac])
		if err != nil {
			c.SendWarnMessage(msg.Context)
			c.sdClient.Log(err.Error())
			c.sendColorErrorToPropertyInspector(msg.Action, msg.Context, err)
			return
		}
		c.sendColorStateToPropertyInspector(msg.Action, msg.Context, color)
	case "kofi":
		_ = open.Start("https://ko-fi.com/P5P23OLT2")
	case "discord":
		_ = open.Start("https://discord.gg/PPVYMeP")
	case "twitch":
		_ = open.Start("https://twitch.tv/wwsean08")
	case "gSettingsHelp":
		version := c.appInfo.Version
		if version == "develop" {
			version = "main"
		}
		site := fmt.Sprintf("https://gitlab.com/wwsean08/lifx-streamdeck/-/blob/%s/docs/global-settings.md",
			version)
		_ = open.Start(site)
	default:
		c.sdClient.Log(fmt.Sprintf("Unknown message type received from Property Inspector, %s", msg.Payload["type"]))
	}
}

func (c *Client) logf(format string, args ...interface{}) {
	if c.sdClient != nil {
		c.sdClient.Log(fmt.Sprintf(format, args...))
	}
}

// OnKeyUp is called when the Stream Deck key is unpressed
func (c *Client) OnKeyUp(msg streamdeck.KeyUpMsg) {
	context := msg.Context
	c.logf("lifx-plus keyUp action=%s", msg.Action)
	switch msg.Action {
	case ActionTurnOnDevice:
		powerSettings := new(models.PowerSettings)
		err := mapstructure.Decode(msg.Payload.Settings, powerSettings)
		if err != nil {
			c.sdClient.Log(err.Error())
			c.SendWarnMessage(context)
			return
		}
		c.logf("lifx-plus power decoded action=%s devices=%d transition=%d", msg.Action, len(powerSettings.Devices), powerSettings.Transition)
		err = c.controller.SetPowerState(powerSettings, true)
		if err != nil {
			c.sdClient.Log(err.Error())
		}
	case ActionTurnOffDevice:
		powerSettings := new(models.PowerSettings)
		err := mapstructure.Decode(msg.Payload.Settings, powerSettings)
		if err != nil {
			c.sdClient.Log(err.Error())
			c.SendWarnMessage(context)
			return
		}
		c.logf("lifx-plus power decoded action=%s devices=%d transition=%d", msg.Action, len(powerSettings.Devices), powerSettings.Transition)
		err = c.controller.SetPowerState(powerSettings, false)
		if err != nil {
			c.sdClient.Log(err.Error())
		}
	case ActionSetColor:
		colorSettings := new(models.ColorSettings)
		err := mapstructure.Decode(msg.Payload.Settings, colorSettings)
		if err != nil {
			c.sdClient.Log(err.Error())
			c.SendWarnMessage(context)
			return
		}
		err = c.controller.SetColor(colorSettings)
		if err != nil {
			c.sdClient.Log(err.Error())
		}
	case ActionSetBrightness:
		brightnessSettings := new(models.BrightnessSettings)
		err := mapstructure.Decode(msg.Payload.Settings, brightnessSettings)
		if err != nil {
			c.sdClient.Log(err.Error())
			c.SendWarnMessage(context)
			return
		}
		err = c.controller.SetBrightness(brightnessSettings)
		if err != nil {
			c.sdClient.Log(err.Error())
		}
	case ActionSetWaveform:
		waveFormSettings := new(models.WaveFormSettings)
		err := mapstructure.Decode(msg.Payload.Settings, waveFormSettings)
		if err != nil {
			c.sdClient.Log(err.Error())
			c.SendWarnMessage(context)
			return
		}
		err = c.controller.SetWaveform(waveFormSettings)
		if err != nil {
			c.sdClient.Log(err.Error())
		}
	case ActionAnimatedScene:
		sceneSettings := new(models.SceneSettings)
		err := mapstructure.Decode(msg.Payload.Settings, sceneSettings)
		if err != nil {
			c.sdClient.Log(err.Error())
			c.SendWarnMessage(context)
			return
		}
		c.logf("lifx-plus scene decoded devices=%d preset=%s reverse=%v duration=%d stagger=%d", len(sceneSettings.Devices), sceneSettings.Preset, sceneSettings.Reverse, sceneSettings.Duration, sceneSettings.Stagger)
		err = c.controller.SetScene(sceneSettings)
		if err != nil {
			c.sdClient.Log(err.Error())
		}
	case ActionToggleDevice:
		toggleSettings := new(models.ToggleSettings)
		err := mapstructure.Decode(msg.Payload.Settings, toggleSettings)
		if err != nil {
			c.sdClient.Log(err.Error())
			c.SendWarnMessage(context)
			return
		}
		err = c.controller.TogglePowerState(toggleSettings)
		if err != nil {
			c.sdClient.Log(err.Error())
		}
	case ActionDebug:
		debug, err := c.controller.Debug()
		if err != nil {
			c.sdClient.Log(err.Error())
			c.SendWarnMessage(context)
			return
		}
		c.generateDebug(msg.Context, debug)
	default:
		c.SendWarnMessage(msg.Context)
		c.sdClient.Log(fmt.Sprintf("Unknown action received %s", msg.Action))
	}
}

func (c *Client) OnDidReceiveGlobalSettings(msg streamdeck.DidReceiveGlobalSettingsMsg) {
	settings := msg.Payload.Settings
	// generally this will be a no-op but makes logic simpler
	migratedSettings, err := MigrateGlobalSettings(settings)
	if err != nil {
		c.sdClient.Log(err.Error())
		return
	}
	update := streamdeck.SetGlobalSettingsMsg{
		Event:   streamdeck.SetGlobalSettingsEvent,
		Context: c.uuid,
		Payload: migratedSettings,
	}
	_ = c.sdClient.SendMessage(update)
	c.UpdateGlobalSettings(migratedSettings)
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

func (c *Client) sendColorErrorToPropertyInspector(action, context string, err error) {
	msg := streamdeck.SendToPropertyInspectorMsg{
		Action:  action,
		Context: context,
		Event:   streamdeck.SendToPropertyInspectorEvent,
		Payload: map[string]interface{}{
			"type":  "getColorError",
			"error": err.Error(),
		},
	}
	sendErr := c.sdClient.SendMessage(msg)
	if sendErr != nil {
		c.sdClient.Log(sendErr.Error())
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
			label = mac
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

func (c *Client) UpdateGlobalSettings(settings map[string]interface{}) {
	var gSettings = new(models.GlobalSettings)
	err := mapstructure.Decode(settings, gSettings)
	if err != nil {
		c.sdClient.Log(err.Error())
		return
	}

	if c.scheduler == nil {
		c.scheduler = quartz.NewStdScheduler()
		c.scheduler.Start()
	}
	rediscover := RediscoverJob{controller: c.controller}
	// job may not exist and that's fine
	_ = c.scheduler.DeleteJob(rediscover.Key())
	_ = c.scheduler.ScheduleJob(rediscover, quartz.NewSimpleTrigger(time.Hour))

	*c.globalSettings = *gSettings

	_ = c.controller.UpdateGlobalSettings(gSettings)
}
