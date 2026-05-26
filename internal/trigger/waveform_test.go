package trigger

import (
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/wwsean08/lifx-streamdeck/models"
)

func TestDecodeWaveformSettingsUsesStreamDeckLabelAsName(t *testing.T) {
	raw := map[string]interface{}{
		"version":     float64(1),
		"isTransient": true,
		"period":      float64(300),
		"cycles":      float64(1),
		"skew":        float64(-90),
		"waveform":    float64(1),
		"color": map[string]interface{}{
			"hue":        float64(260),
			"saturation": float64(100),
			"brightness": float64(100),
			"kelvin":     float64(5000),
		},
		"devices": []interface{}{
			map[string]interface{}{"label": "Right Lamp", "mac": "d0:73:d5:01:be:91"},
		},
	}

	settings, err := DecodeWaveformSettings(raw)
	require.NoError(t, err)
	require.Len(t, settings.Devices, 1)
	require.Equal(t, "Right Lamp", settings.Devices[0].Name)
	require.Equal(t, "d0:73:d5:01:be:91", settings.Devices[0].Mac)
	require.True(t, settings.IsTransient)
	require.Equal(t, uint32(300), settings.Period)
	require.Equal(t, float32(1), settings.Cycles)
	require.Equal(t, int16(-90), settings.Skew)
	require.Equal(t, uint8(1), settings.Waveform)
	require.Equal(t, uint(260), settings.AlternateColor.Hue)
}

func TestBuildWaveformPacketTargetsMacAndUsesWaveformPayload(t *testing.T) {
	settings := mustDecodeWaveform(t, map[string]interface{}{
		"version":     float64(1),
		"isTransient": true,
		"period":      float64(300),
		"cycles":      float64(1),
		"skew":        float64(-90),
		"waveform":    float64(1),
		"color": map[string]interface{}{
			"hue":        float64(260),
			"saturation": float64(100),
			"brightness": float64(100),
			"kelvin":     float64(5000),
		},
		"devices": []interface{}{
			map[string]interface{}{"label": "Right Lamp", "mac": "d0:73:d5:01:be:91"},
		},
	})

	packet, err := buildWaveformPacket(settings, "d0:73:d5:01:be:91")
	require.NoError(t, err)
	require.Len(t, packet, 57)
	require.Equal(t, uint16(57), binary.LittleEndian.Uint16(packet[0:2]))
	require.Equal(t, []byte{0xd0, 0x73, 0xd5, 0x01, 0xbe, 0x91, 0x00, 0x00}, packet[8:16])
	require.Equal(t, byte(1), packet[22])
	require.Equal(t, uint16(lifxSetWaveformType), binary.LittleEndian.Uint16(packet[32:34]))
	require.Equal(t, byte(1), packet[37])
	require.Equal(t, uint32(300), binary.LittleEndian.Uint32(packet[46:50]))
	require.Equal(t, uint16(0x8d0a), binary.LittleEndian.Uint16(packet[54:56]))
	require.Equal(t, byte(1), packet[56])
}

func TestFindWaveformSettingsByTitle(t *testing.T) {
	streamDeckDir := t.TempDir()
	profileDir := filepath.Join(streamDeckDir, "ProfilesV3", "device.sdProfile", "Profiles", "profile")
	require.NoError(t, os.MkdirAll(profileDir, 0755))
	manifestPath := filepath.Join(profileDir, "manifest.json")
	manifest := map[string]interface{}{
		"Controllers": []interface{}{
			map[string]interface{}{
				"Actions": map[string]interface{}{
					"4,2": map[string]interface{}{
						"UUID": WaveformActionUUID,
						"Settings": map[string]interface{}{
							"version":     1,
							"isTransient": true,
							"period":      300,
							"cycles":      1,
							"skew":        -90,
							"waveform":    1,
							"color": map[string]interface{}{
								"hue":        260,
								"saturation": 100,
								"brightness": 100,
								"kelvin":     5000,
							},
							"devices": []interface{}{
								map[string]interface{}{"label": "Desk Lamp", "mac": "d0:73:d5:02:03:94"},
							},
						},
						"States": []interface{}{
							map[string]interface{}{"Title": "Codex Done"},
						},
					},
				},
			},
		},
	}
	writeJSON(t, manifestPath, manifest)

	settings, path, err := FindWaveformSettingsByTitle(streamDeckDir, "Codex Done")
	require.NoError(t, err)
	require.Equal(t, manifestPath, path)
	require.Len(t, settings.Devices, 1)
	require.Equal(t, "Desk Lamp", settings.Devices[0].Name)
	require.Equal(t, uint(260), settings.AlternateColor.Hue)
}

func writeJSON(t *testing.T, path string, value interface{}) {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	require.NoError(t, ioutil.WriteFile(path, data, 0644))
}

func mustDecodeWaveform(t *testing.T, raw map[string]interface{}) *models.WaveFormSettings {
	t.Helper()
	settings, err := DecodeWaveformSettings(raw)
	require.NoError(t, err)
	return settings
}
