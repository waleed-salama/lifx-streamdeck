package trigger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/mitchellh/mapstructure"
	"gitlab.com/wwsean08/lifx-streamdeck/models"
)

const WaveformActionUUID = "com.waleed-salama.lifx-plus.waveform"

func DefaultStreamDeckDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "com.elgato.StreamDeck"), nil
}

func LoadWaveformSettings(path string) (*models.WaveFormSettings, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}
	return DecodeWaveformSettings(settings)
}

func FindWaveformSettingsByTitle(streamDeckDir, title string) (*models.WaveFormSettings, string, error) {
	paths, err := profileManifestPaths(streamDeckDir)
	if err != nil {
		return nil, "", err
	}

	for _, path := range paths {
		settings, ok, err := waveformSettingsInManifest(path, title)
		if err != nil {
			return nil, "", err
		}
		if ok {
			return settings, path, nil
		}
	}
	return nil, "", fmt.Errorf("waveform action titled %q not found under %s", title, streamDeckDir)
}

func DecodeWaveformSettings(raw map[string]interface{}) (*models.WaveFormSettings, error) {
	normalizeDeviceLabels(raw)
	settings := new(models.WaveFormSettings)
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  settings,
		TagName: "mapstructure",
	})
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(raw); err != nil {
		return nil, err
	}
	if len(settings.Devices) == 0 {
		return nil, fmt.Errorf("waveform settings contain no devices")
	}
	return settings, nil
}

func profileManifestPaths(streamDeckDir string) ([]string, error) {
	var paths []string
	for _, profileDir := range []string{"ProfilesV3", "ProfilesV2"} {
		root := filepath.Join(streamDeckDir, profileDir)
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info == nil || info.IsDir() || info.Name() != "manifest.json" {
				return nil
			}
			paths = append(paths, path)
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	sort.Strings(paths)
	return paths, nil
}

func waveformSettingsInManifest(path, title string) (*models.WaveFormSettings, bool, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, false, err
	}
	var manifest interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, false, err
	}
	raw, ok := findWaveformSettings(manifest, title)
	if !ok {
		return nil, false, nil
	}
	settings, err := DecodeWaveformSettings(raw)
	if err != nil {
		return nil, false, err
	}
	return settings, true, nil
}

func findWaveformSettings(value interface{}, title string) (map[string]interface{}, bool) {
	switch typed := value.(type) {
	case map[string]interface{}:
		if typed["UUID"] == WaveformActionUUID && actionHasTitle(typed, title) {
			if settings, ok := typed["Settings"].(map[string]interface{}); ok {
				return settings, true
			}
		}
		for _, child := range typed {
			if settings, ok := findWaveformSettings(child, title); ok {
				return settings, true
			}
		}
	case []interface{}:
		for _, child := range typed {
			if settings, ok := findWaveformSettings(child, title); ok {
				return settings, true
			}
		}
	}
	return nil, false
}

func actionHasTitle(action map[string]interface{}, title string) bool {
	states, ok := action["States"].([]interface{})
	if !ok {
		return false
	}
	for _, state := range states {
		stateMap, ok := state.(map[string]interface{})
		if !ok {
			continue
		}
		if stateTitle, ok := stateMap["Title"].(string); ok && stateTitle == title {
			return true
		}
	}
	return false
}

func normalizeDeviceLabels(raw map[string]interface{}) {
	devices, ok := raw["devices"].([]interface{})
	if !ok {
		return
	}
	for _, device := range devices {
		deviceMap, ok := device.(map[string]interface{})
		if !ok {
			continue
		}
		if _, ok := deviceMap["name"]; ok {
			continue
		}
		if label, ok := deviceMap["label"]; ok {
			deviceMap["name"] = label
		}
	}
}

func DeviceName(device models.Device) string {
	return deviceName(device)
}

func deviceName(device models.Device) string {
	if device.Name != "" {
		return fmt.Sprintf("%s (%s)", device.Name, device.Mac)
	}
	return device.Mac
}
