package streamdeck

import (
	"github.com/stretchr/testify/mock"
	"gitlab.com/wwsean08/lifx-streamdeck/mocks"
	"gitlab.com/wwsean08/lifx-streamdeck/models"
	"gitlab.com/wwsean08/streamdeck"
	"testing"
)

func TestClient_TurnOnDevice(t *testing.T) {
	data := map[string]interface{}{
		"version":    2,
		"transition": 0,
		"devices": []models.Device{
			{
				Name: "foo",
				Mac:  "bar",
			},
		},
	}
	testController := new(mocks.Controller)
	testController.On("SetPowerState", mock.AnythingOfType("*models.PowerSettings"), true).Return(nil)

	client := new(Client)
	client.controller = testController
	msg := streamdeck.KeyUpMsg{
		Action: ActionTurnOnDevice,
		Payload: struct {
			Settings    map[string]interface{} `json:"settings"`
			Coordinates streamdeck.Coordinates `json:"coordinates"`
		}(struct {
			Settings    map[string]interface{}
			Coordinates streamdeck.Coordinates
		}{Settings: data, Coordinates: streamdeck.Coordinates(struct {
			Column uint
			Row    uint
		}{Column: 0, Row: 0})}),
	}
	client.OnKeyUp(msg)

	testController.AssertExpectations(t)
}

func TestClient_TurnOffDevice(t *testing.T) {
	data := map[string]interface{}{
		"version":    2,
		"transition": 0,
		"devices": []models.Device{
			{
				Name: "foo",
				Mac:  "bar",
			},
		},
	}
	testController := new(mocks.Controller)
	testController.On("SetPowerState", mock.AnythingOfType("*models.PowerSettings"), false).Return(nil)

	client := new(Client)
	client.controller = testController
	msg := streamdeck.KeyUpMsg{
		Action: ActionTurnOffDevice,
		Payload: struct {
			Settings    map[string]interface{} `json:"settings"`
			Coordinates streamdeck.Coordinates `json:"coordinates"`
		}(struct {
			Settings    map[string]interface{}
			Coordinates streamdeck.Coordinates
		}{Settings: data, Coordinates: streamdeck.Coordinates(struct {
			Column uint
			Row    uint
		}{Column: 0, Row: 0})}),
	}
	client.OnKeyUp(msg)

	testController.AssertExpectations(t)
}

func TestClient_SetColor(t *testing.T) {
	data := map[string]interface{}{
		"version": 2,
		"color": models.Color{
			Hue:        0,
			Saturation: 0,
			Brightness: 0,
			Kelvin:     5000,
			Transition: 0,
		},
		"devices": []models.Device{
			{
				Name: "foo",
				Mac:  "bar",
			},
		},
	}
	testController := new(mocks.Controller)
	testController.On("SetColor", mock.AnythingOfType("*models.ColorSettings")).Return(nil)

	client := new(Client)
	client.controller = testController
	msg := streamdeck.KeyUpMsg{
		Action: ActionSetColor,
		Payload: struct {
			Settings    map[string]interface{} `json:"settings"`
			Coordinates streamdeck.Coordinates `json:"coordinates"`
		}(struct {
			Settings    map[string]interface{}
			Coordinates streamdeck.Coordinates
		}{Settings: data, Coordinates: streamdeck.Coordinates(struct {
			Column uint
			Row    uint
		}{Column: 0, Row: 0})}),
	}
	client.OnKeyUp(msg)

	testController.AssertExpectations(t)
}

func TestClient_SetBrightness(t *testing.T) {
	data := map[string]interface{}{
		"version":    2,
		"transition": 0,
		"brightness": 0,
		"devices": []models.Device{
			{
				Name: "foo",
				Mac:  "bar",
			},
		},
	}
	testController := new(mocks.Controller)
	testController.On("SetBrightness", mock.AnythingOfType("*models.BrightnessSettings")).Return(nil)

	client := new(Client)
	client.controller = testController
	msg := streamdeck.KeyUpMsg{
		Action: ActionSetBrightness,
		Payload: struct {
			Settings    map[string]interface{} `json:"settings"`
			Coordinates streamdeck.Coordinates `json:"coordinates"`
		}(struct {
			Settings    map[string]interface{}
			Coordinates streamdeck.Coordinates
		}{Settings: data, Coordinates: streamdeck.Coordinates(struct {
			Column uint
			Row    uint
		}{Column: 0, Row: 0})}),
	}
	client.OnKeyUp(msg)

	testController.AssertExpectations(t)
}

func TestClient_Waveform(t *testing.T) {
	data := map[string]interface{}{
		"version":     2,
		"isTransient": true,
		"period":      1000,
		"skew":        0,
		"waveform":    1,
		"color": models.Color{
			Hue:        0,
			Saturation: 0,
			Brightness: 0,
			Kelvin:     5000,
		},
		"devices": []models.Device{
			{
				Name: "foo",
				Mac:  "bar",
			},
		},
	}
	testController := new(mocks.Controller)
	testController.On("SetWaveform", mock.AnythingOfType("*models.WaveFormSettings")).Return(nil)

	client := new(Client)
	client.controller = testController
	msg := streamdeck.KeyUpMsg{
		Action: ActionSetWaveform,
		Payload: struct {
			Settings    map[string]interface{} `json:"settings"`
			Coordinates streamdeck.Coordinates `json:"coordinates"`
		}(struct {
			Settings    map[string]interface{}
			Coordinates streamdeck.Coordinates
		}{Settings: data, Coordinates: streamdeck.Coordinates(struct {
			Column uint
			Row    uint
		}{Column: 0, Row: 0})}),
	}
	client.OnKeyUp(msg)

	testController.AssertExpectations(t)
}

func TestClient_Toggle(t *testing.T) {
	data := map[string]interface{}{
		"version":    2,
		"transition": 0,
		"devices": []models.Device{
			{
				Name: "foo",
				Mac:  "bar",
			},
		},
	}
	testController := new(mocks.Controller)
	testController.On("TogglePowerState", mock.AnythingOfType("*models.ToggleSettings")).Return(nil)

	client := new(Client)
	client.controller = testController
	msg := streamdeck.KeyUpMsg{
		Action: ActionToggleDevice,
		Payload: struct {
			Settings    map[string]interface{} `json:"settings"`
			Coordinates streamdeck.Coordinates `json:"coordinates"`
		}(struct {
			Settings    map[string]interface{}
			Coordinates streamdeck.Coordinates
		}{Settings: data, Coordinates: streamdeck.Coordinates(struct {
			Column uint
			Row    uint
		}{Column: 0, Row: 0})}),
	}
	client.OnKeyUp(msg)

	testController.AssertExpectations(t)
}
