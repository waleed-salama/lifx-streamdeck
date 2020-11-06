package lifx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigratePowerFromUnversionedToCurrent(t *testing.T) {
	// v0 power settings: {"d0:73:d5:2b:a7:b8":"Beam"}
	settings := make(map[string]interface{})
	settings["d0:73:d5:2b:a7:b8"] = "Beam"

	require.Equal(t, "Beam", settings["d0:73:d5:2b:a7:b8"])

	settings = migratePowerSettingsToV1(settings)
	require.Equal(t, 1, settings["version"])
	devices := settings["devices"].([]map[string]interface{})
	require.Equal(t, "d0:73:d5:2b:a7:b8", devices[0]["mac"])
	require.Equal(t, "Beam", devices[0]["name"])
}

func TestMigrateColorSettingsFromUnversionedToCurrent(t *testing.T) {
	// v0 color settings {"color":{"brightness":"100","hue":"90","kelvin":"5000","saturation":"100"},"devices":{"d0:73:d5:2b:a7:b8":"Beam","d0:73:d5:3c:86:05":"Office 1","d0:73:d5:3c:dc:6d":"Office 2"}}
	settings := map[string]interface{}{
		"color": map[string]interface{}{
			"hue":        "90",
			"saturation": "100",
			"brightness": "100",
			"kelvin":     "5000",
		},
		"devices": map[string]interface{}{
			"d0:73:d5:2b:a7:b8": "Beam",
			"d0:73:d5:3c:86:05": "Office 1",
			"d0:73:d5:3c:dc:6d": "Office 2",
		},
	}

	require.Len(t, settings["devices"], 3)
	require.Len(t, settings["color"], 4)
	require.NotContains(t, settings, "version")

	settings = migrateColorSettingsToV1(settings)
	require.Equal(t, 1, settings["version"])
	colorSettings := settings["color"].(map[string]interface{})
	require.Equal(t, "0", colorSettings["transition"])

	settings, err := migrateColorSettingsToV2(settings)
	require.NoError(t, err)
	colorSettings = settings["color"].(map[string]interface{})
	deviceSettings := settings["devices"].([]map[string]interface{})
	require.NotNil(t, colorSettings)
	require.NotNil(t, deviceSettings)

	require.Equal(t, uint64(90), colorSettings["hue"])
	require.Equal(t, uint64(100), colorSettings["saturation"])
	require.Equal(t, uint64(100), colorSettings["brightness"])
	require.Equal(t, uint64(5000), colorSettings["kelvin"])
	require.Equal(t, uint64(0), colorSettings["transition"])

	_a7b8FoundTimes := 0
	_8605FoundTimes := 0
	_dc6dFoundTimes := 0
	for _, setting := range deviceSettings {
		switch setting["mac"] {
		case "d0:73:d5:2b:a7:b8":
			require.Equal(t, "Beam", setting["name"])
			_a7b8FoundTimes++
		case "d0:73:d5:3c:86:05":
			require.Equal(t, "Office 1", setting["name"])
			_8605FoundTimes++
		case "d0:73:d5:3c:dc:6d":
			require.Equal(t, "Office 2", setting["name"])
			_dc6dFoundTimes++
		}
	}
	require.Equal(t, 1, _a7b8FoundTimes)
	require.Equal(t, 1, _8605FoundTimes)
	require.Equal(t, 1, _dc6dFoundTimes)
}

func TestMigrateBrightnessSettingsFromUnversionedToCurrent(t *testing.T) {
	// v0 power settings: {"brightness":"50","devices":{"d0:73:d5:2b:a7:b8":"Beam","d0:73:d5:3c:dc:6d":"Office 2"}}
	settings := map[string]interface{}{
		"brightness": "50",
		"devices": map[string]interface{}{
			"d0:73:d5:2b:a7:b8": "Beam",
			"d0:73:d5:3c:dc:6d": "Office 2",
		},
	}
	require.Equal(t, "50", settings["brightness"])
	require.Len(t, settings["devices"], 2)
	require.NotContains(t, settings, "version")

	settings = migrateBrightnessSettingsToV1(settings)
	require.Equal(t, 1, settings["version"])
	require.Equal(t, "0", settings["transition"])

	settings, err := migrateBrightnessSettingsToV2(settings)
	require.NoError(t, err)
	require.Equal(t, 2, settings["version"])
	require.Equal(t, uint64(0), settings["transition"])
	require.Equal(t, uint64(50), settings["brightness"])
	devices := settings["devices"].([]map[string]interface{})

	_a7b8FoundTimes := 0
	_dc6dFoundTimes := 0
	for _, setting := range devices {
		switch setting["mac"] {
		case "d0:73:d5:2b:a7:b8":
			require.Equal(t, "Beam", setting["name"])
			_a7b8FoundTimes++
		case "d0:73:d5:3c:dc:6d":
			require.Equal(t, "Office 2", setting["name"])
			_dc6dFoundTimes++
		}
	}
	require.Equal(t, 1, _a7b8FoundTimes)
	require.Equal(t, 1, _dc6dFoundTimes)
}

func TestMigrateWaveformSettingsFromV1ToCurrent(t *testing.T) {
	// there are no migrations
}

func TestMigrate(t *testing.T) {
	require := require.New(t)
	testCases := []struct {
		desc             string
		client           *Client
		settings         map[string]interface{}
		expectedSettings map[string]interface{}
		action           string
		expectErr        bool
	}{
		{
			desc:   "Test Migrate v0 to current color settings",
			client: &Client{},
			settings: map[string]interface{}{
				"color": map[string]interface{}{
					"hue":        "90",
					"saturation": "100",
					"brightness": "100",
					"kelvin":     "5000",
				},
				"devices": map[string]interface{}{
					"d0:73:d5:2b:a7:b8": "Beam",
					"d0:73:d5:3c:86:05": "Office 1",
					"d0:73:d5:3c:dc:6d": "Office 2",
				},
			},
			expectedSettings: map[string]interface{}{
				"version": 2,
				"color": map[string]interface{}{
					"hue":        uint64(90),
					"saturation": uint64(100),
					"brightness": uint64(100),
					"kelvin":     uint64(5000),
					"transition": uint64(0),
				},
				"devices": []map[string]interface{}{
					{"mac": "d0:73:d5:2b:a7:b8", "name": "Beam"},
					{"mac": "d0:73:d5:3c:86:05", "name": "Office 1"},
					{"mac": "d0:73:d5:3c:dc:6d", "name": "Office 2"},
				},
			},
			action:    ActionSetColor,
			expectErr: false,
		},
		{
			desc:   "Test Migrate v0 to current brightness settings",
			client: &Client{},
			settings: map[string]interface{}{
				"brightness": "50",
				"devices": map[string]interface{}{
					"d0:73:d5:2b:a7:b8": "Beam",
					"d0:73:d5:3c:dc:6d": "Office 2",
				},
			},
			expectedSettings: map[string]interface{}{
				"version":    2,
				"brightness": uint64(50),
				"transition": uint64(0),
				"devices": []map[string]interface{}{
					{"mac": "d0:73:d5:2b:a7:b8", "name": "Beam"},
					{"mac": "d0:73:d5:3c:dc:6d", "name": "Office 2"},
				},
			},
			action:    ActionSetBrightness,
			expectErr: false,
		},
		{
			desc:   "Test Migrate v0 to current power settings",
			client: &Client{},
			settings: map[string]interface{}{
				"d0:73:d5:2b:a7:b8": "Beam",
			},
			expectedSettings: map[string]interface{}{
				"version": 1,
				"devices": []map[string]interface{}{
					{"mac": "d0:73:d5:2b:a7:b8", "name": "Beam"},
				},
			},
			action:    ActionTurnOffDevice,
			expectErr: false,
		},
		{
			desc:   "Test Migrate v0 to current power settings",
			client: &Client{},
			settings: map[string]interface{}{
				"d0:73:d5:2b:a7:b8": "Beam",
			},
			expectedSettings: map[string]interface{}{
				"version": 1,
				"devices": []map[string]interface{}{
					{"mac": "d0:73:d5:2b:a7:b8", "name": "Beam"},
				},
			},
			action:    ActionTurnOnDevice,
			expectErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			settings, err := tC.client.Migrate(tC.settings, tC.action)
			if tC.expectErr {
				require.Error(err)
			} else {
				require.NoError(err)
			}
			require.Equal(tC.expectedSettings, settings)
		})
	}
}
