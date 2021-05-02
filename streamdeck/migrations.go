package streamdeck

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

//MigrateActions migrates the schema of settings objects
func (c *Client) MigrateActions(settings map[string]interface{}, action string) (map[string]interface{}, error) {
	switch action {
	case ActionSetColor:
		if val, ok := settings["version"]; ok {
			intVal, err := versionToInt(val)
			if err != nil {
				return settings, err
			}
			switch intVal {
			case 1:
				settings, err = migrateColorSettingsToV2(settings)
				if err != nil {
					return settings, err
				}
				fallthrough
			case 2:

			}
		} else {
			settings = migrateColorSettingsToV1(settings)
			return c.MigrateActions(settings, action)
		}
	case ActionSetBrightness:
		if val, ok := settings["version"]; ok {
			intVal, err := versionToInt(val)
			if err != nil {
				return settings, err
			}
			switch intVal {
			case 1:
				settings, err = migrateBrightnessSettingsToV2(settings)
				if err != nil {
					return settings, err
				}
				fallthrough
			case 2:

			}
		} else {
			settings = migrateBrightnessSettingsToV1(settings)
			return c.MigrateActions(settings, action)
		}
	case ActionTurnOnDevice, ActionTurnOffDevice:
		if val, ok := settings["version"]; ok {
			intVal, err := versionToInt(val)
			if err != nil {
				return settings, err
			}
			switch intVal {
			case 1:
				settings = migratePowerSettingsToV2(settings)
				fallthrough
			case 2:
				// doesn't exist yet
			}
		} else {
			settings = migratePowerSettingsToV1(settings)
			return c.MigrateActions(settings, action)
		}
	case ActionToggleDevice:
		if val, ok := settings["version"]; ok {
			intVal, err := versionToInt(val)
			if err != nil {
				return settings, err
			}
			switch intVal {
			// no migrations for this yet
			}
		}
	case ActionSetWaveform:
		if val, ok := settings["version"]; ok {
			intVal, err := versionToInt(val)
			if err != nil {
				return settings, err
			}
			switch intVal {
			// no migrations for this yet
			}
		}
	}
	return settings, nil
}

func MigrateGlobalSettings(globalSettings map[string]interface{}) (map[string]interface{}, error) {
	// No settings, migrate to v1
	var err error
	if len(globalSettings) == 0 {
		globalSettings["version"] = 1
		globalSettings["outIP"] = ""
		globalSettings, err = MigrateGlobalSettings(globalSettings)
	} else {
		if val, ok := globalSettings["version"]; ok {
			intVal, err := versionToInt(val)
			if err != nil {
				return globalSettings, err
			}
			switch intVal {
			case 1:
				globalSettings["version"] = 2
				// if directComm doesn't exist, add it
				if _, ok := globalSettings["directComm"]; !ok {
					globalSettings["directComm"] = false
				}
				fallthrough
			case 2:
				globalSettings["version"] = 3
				if _, ok := globalSettings["cacheTTL"]; !ok {
					globalSettings["cacheTTL"] = time.Minute
				}
				fallthrough
			case 3:
				globalSettings["version"] = 4
				delete(globalSettings, "cacheTTL")
				globalSettings["customDevices"] = []string{}
			}
		}
	}
	return globalSettings, err
}

func migrateColorSettingsToV1(settings map[string]interface{}) map[string]interface{} {
	settings["version"] = 1
	colorSettings := settings["color"].(map[string]interface{})
	colorSettings["transition"] = "0"
	return settings
}

func migrateColorSettingsToV2(settings map[string]interface{}) (map[string]interface{}, error) {
	// handle colors
	colorSettings := settings["color"].(map[string]interface{})
	deviceSettings := settings["devices"].(map[string]interface{})
	hue, err := strconv.ParseUint(colorSettings["hue"].(string), 10, 16)
	if err != nil {
		return settings, err
	}
	saturation, err := strconv.ParseUint(colorSettings["saturation"].(string), 10, 16)
	if err != nil {
		return settings, err
	}
	brightness, err := strconv.ParseUint(colorSettings["brightness"].(string), 10, 16)
	if err != nil {
		return settings, err
	}
	kelvin, err := strconv.ParseUint(colorSettings["kelvin"].(string), 10, 16)
	if err != nil {
		return settings, err
	}
	transition, err := strconv.ParseUint(colorSettings["transition"].(string), 10, 16)
	if err != nil {
		return settings, err
	}
	// end handle colorsDeviceinterface{}))
	devices := make([]map[string]interface{}, 0, len(deviceSettings))
	for key, val := range deviceSettings {
		tmpDevice := map[string]interface{}{"mac": key, "name": val}
		devices = append(devices, tmpDevice)
	}
	//end handle devices

	colorSettings["hue"] = hue
	colorSettings["saturation"] = saturation
	colorSettings["brightness"] = brightness
	colorSettings["kelvin"] = kelvin
	colorSettings["transition"] = transition
	settings["color"] = colorSettings
	settings["devices"] = devices
	settings["version"] = 2
	return settings, nil
}

func migrateBrightnessSettingsToV1(settings map[string]interface{}) map[string]interface{} {
	settings["version"] = 1
	settings["transition"] = "0"
	return settings
}

func migrateBrightnessSettingsToV2(settings map[string]interface{}) (map[string]interface{}, error) {
	transition, err := strconv.ParseUint(settings["transition"].(string), 10, 16)
	if err != nil {
		return settings, err
	}
	brightness, err := strconv.ParseUint(settings["brightness"].(string), 10, 16)
	if err != nil {
		return settings, err
	}
	deviceSettings := settings["devices"].(map[string]interface{})
	devices := make([]map[string]interface{}, 0, len(deviceSettings))
	for key, val := range deviceSettings {
		tmpDevice := map[string]interface{}{"mac": key, "name": val}
		devices = append(devices, tmpDevice)
	}

	settings["transition"] = transition
	settings["brightness"] = brightness
	settings["devices"] = devices
	settings["version"] = 2
	return settings, nil
}

func migratePowerSettingsToV1(settings map[string]interface{}) map[string]interface{} {
	devices := make([]map[string]interface{}, 0, len(settings))
	for key, val := range settings {
		tmpDevice := map[string]interface{}{"mac": key, "name": val}
		devices = append(devices, tmpDevice)
	}

	newSettings := make(map[string]interface{})
	newSettings["version"] = 1
	newSettings["devices"] = devices
	return newSettings
}

func migratePowerSettingsToV2(settings map[string]interface{}) map[string]interface{} {
	settings["version"] = 2
	settings["transition"] = 0
	return settings
}

// versionToInt is here to make sure that the switch statements work
func versionToInt(version interface{}) (int, error) {
	switch version.(type) {
	case int:
		return version.(int), nil
	case float64:
		return int(version.(float64)), nil
	case float32:
		return int(version.(float32)), nil
	case int8:
		return int(version.(int8)), nil
	case int16:
		return int(version.(int16)), nil
	case int32:
		return int(version.(int32)), nil
	case int64:
		return int(version.(int64)), nil
	}
	return -1, fmt.Errorf("Error converting version to int, current type %s\n", reflect.TypeOf(version))
}
