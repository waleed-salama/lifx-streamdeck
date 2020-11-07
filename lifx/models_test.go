package lifx

import (
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
	"gitlab.com/wwsean08/golifx"
	"testing"
)

func TestPowerSettingsDecode(t *testing.T) {
	settings := map[string]interface{}{
		"version": 1,
		"devices": []map[string]interface{}{
			{
				"mac":  "d0:73:d5:2b:a7:b8",
				"name": "beam",
			},
		},
	}

	powerSettings := new(PowerSettings)
	err := mapstructure.Decode(settings, powerSettings)
	require.NoError(t, err)
	require.Equal(t, 1, powerSettings.Version)
	require.Len(t, powerSettings.Devices, 1)
	require.Equal(t, "d0:73:d5:2b:a7:b8", powerSettings.Devices[0].Mac)
	require.Equal(t, "beam", powerSettings.Devices[0].Name)
}

func TestToggleSettingsDecode(t *testing.T) {
	settings := map[string]interface{}{
		"version": 1,
		"devices": []map[string]interface{}{
			{
				"mac":  "d0:73:d5:2b:a7:b8",
				"name": "beam",
			},
		},
	}

	toggleSettings := new(ToggleSettings)
	err := mapstructure.Decode(settings, toggleSettings)
	require.NoError(t, err)
	require.Equal(t, 1, toggleSettings.Version)
	require.Len(t, toggleSettings.Devices, 1)
	require.Equal(t, "d0:73:d5:2b:a7:b8", toggleSettings.Devices[0].Mac)
	require.Equal(t, "beam", toggleSettings.Devices[0].Name)
}

func TestColorSettingsDecode(t *testing.T) {
	settings := map[string]interface{}{
		"version": 2,
		"color": map[string]interface{}{
			"hue":        180,
			"saturation": 100,
			"brightness": 100,
			"kelvin":     5000,
			"transition": 5000,
		},
		"devices": []map[string]interface{}{
			{
				"mac":  "d0:73:d5:2b:a7:b8",
				"name": "beam",
			},
		},
	}

	colorSettings := new(ColorSettings)
	err := mapstructure.Decode(settings, colorSettings)
	require.NoError(t, err)
	require.Equal(t, 2, colorSettings.Version)
	require.Len(t, colorSettings.Devices, 1)
	require.Equal(t, "d0:73:d5:2b:a7:b8", colorSettings.Devices[0].Mac)
	require.Equal(t, "beam", colorSettings.Devices[0].Name)
	require.Equal(t, uint(180), colorSettings.Color.Hue)
	require.Equal(t, uint(100), colorSettings.Color.Saturation)
	require.Equal(t, uint(100), colorSettings.Color.Brightness)
	require.Equal(t, uint16(5000), colorSettings.Color.Kelvin)
	require.Equal(t, uint32(5000), colorSettings.Color.Transition)

	hue, saturation, brightness, kelvin, transition := colorSettings.Color.GenerateLIFXValues()
	require.Equal(t, uint16(32760), hue)
	require.Equal(t, uint16(65500), saturation)
	require.Equal(t, uint16(65500), brightness)
	require.Equal(t, uint16(5000), kelvin)
	require.Equal(t, uint32(5000), transition)
}

func TestBrightnessSettingsDecode(t *testing.T) {
	settings := map[string]interface{}{
		"version":    2,
		"brightness": 100,
		"transition": 5000,
		"devices": []map[string]interface{}{
			{
				"mac":  "d0:73:d5:2b:a7:b8",
				"name": "beam",
			},
		},
	}

	brightnessSettings := new(BrightnessSettings)
	err := mapstructure.Decode(settings, brightnessSettings)
	require.NoError(t, err)
	require.Equal(t, 2, brightnessSettings.Version)
	require.Len(t, brightnessSettings.Devices, 1)
	require.Equal(t, "d0:73:d5:2b:a7:b8", brightnessSettings.Devices[0].Mac)
	require.Equal(t, "beam", brightnessSettings.Devices[0].Name)
	require.Equal(t, uint(100), brightnessSettings.Brightness)
	require.Equal(t, uint32(5000), brightnessSettings.Transition)
}

func TestWaveFormSettingsDecode(t *testing.T) {
	settings := map[string]interface{}{
		"version":     1,
		"isTransient": false,
		"period":      5000,
		"cycles":      12,
		"skew":        -50,
		"waveform":    golifx.WaveformSine,
		"color": map[string]interface{}{
			"hue":        180,
			"saturation": 100,
			"brightness": 100,
			"kelvin":     5000,
			"transition": 5000,
		},
		"devices": []map[string]interface{}{
			{
				"mac":  "d0:73:d5:2b:a7:b8",
				"name": "beam",
			},
		},
	}

	waveformSettings := new(WaveFormSettings)
	err := mapstructure.Decode(settings, waveformSettings)
	require.NoError(t, err)
	require.Equal(t, 1, waveformSettings.Version)
	require.Len(t, waveformSettings.Devices, 1)
	require.Equal(t, "d0:73:d5:2b:a7:b8", waveformSettings.Devices[0].Mac)
	require.Equal(t, "beam", waveformSettings.Devices[0].Name)
	require.Equal(t, uint(180), waveformSettings.AlternateColor.Hue)
	require.Equal(t, uint(100), waveformSettings.AlternateColor.Saturation)
	require.Equal(t, uint(100), waveformSettings.AlternateColor.Brightness)
	require.Equal(t, uint16(5000), waveformSettings.AlternateColor.Kelvin)
	require.Equal(t, uint32(5000), waveformSettings.AlternateColor.Transition)
	require.Equal(t, golifx.WaveformSine, waveformSettings.Waveform)
	require.Equal(t, uint32(5000), waveformSettings.Period)
	require.Equal(t, float32(12), waveformSettings.Cycles)
	require.Equal(t, int16(-50), waveformSettings.Skew)
	require.False(t, waveformSettings.IsTransient)

	require.Equal(t, int16(-16350), waveformSettings.GetSkewRatio())
}
