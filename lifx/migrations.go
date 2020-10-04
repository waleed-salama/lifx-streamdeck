package lifx

func migrateColorSettingsToV1(settings map[string]interface{}) map[string]interface{} {
	settings["version"] = 1
	colorSettings := settings["color"].(map[string]interface{})
	colorSettings["transition"] = "0"
	return settings
}
