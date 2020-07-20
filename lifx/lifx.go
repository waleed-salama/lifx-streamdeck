package lifx

import (
	"fmt"
	"gitlab.com/wwsean08/golifx"
	"gitlab.com/wwsean08/streamdeck"
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
	}
}

func (c *Client) OnKeyUp(msg streamdeck.KeyUpMsg) {
	switch msg.Action {
	case ActionTurnOnLight:
		println("")
	case ActionTurnOffLight:
		println("")
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

func (c *Client) sendDevicesToPropertyInspector(action string, context string, devices map[string]*golifx.Device) {
	type deviceInfo struct {
		Label string `json:"label"`
		Mac   string `json:"mac"`
	}
	data := []deviceInfo{}
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
