package lifx

import "strconv"

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
