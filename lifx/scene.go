package lifx

import (
	"math"
	"sort"
	"time"

	"gitlab.com/wwsean08/golifx"
	"gitlab.com/wwsean08/lifx-streamdeck/models"
)

const (
	defaultScenePreset      = "arrival"
	defaultSceneDirection   = "left-right"
	defaultSceneColorTravel = "none"
	defaultSceneDuration    = 1600
	maxSceneDuration        = 15000
	maxSceneStagger         = 10000
	maxSceneDevices         = 32
	maxScenePlacements      = 64
	maxSceneTimingOffset    = 30000
	sceneDelayStep          = 50 * time.Millisecond
	sceneOffRestoreDelay    = 750 * time.Millisecond
)

type sceneCommand struct {
	DeviceMac  string
	Delay      time.Duration
	Transition uint32
	Color      *golifx.HSBK
	Power      *bool
}

type sceneTarget struct {
	Color   golifx.HSBK
	PowerOn bool
}

func buildSceneCommands(settings *models.SceneSettings) []sceneCommand {
	if settings == nil || len(settings.Devices) == 0 {
		return nil
	}

	sceneSettings := *settings
	reverse := sceneSettings.Reverse || sceneSettings.Preset == "depart"
	if sceneSettings.Preset == "depart" {
		sceneSettings.Preset = defaultScenePreset
	}

	duration := settings.Duration
	if duration == 0 {
		duration = defaultSceneDuration
	}
	stagger := uint32(float64(settings.Stagger) * (0.5 + clampFloat(float64(settings.Intensity), 0, 100)/100))

	placements := placementMap(&sceneSettings)
	commands := make([]sceneCommand, 0, len(settings.Devices)*4)
	devices := make([]computedSceneDevice, 0, len(settings.Devices))
	maxDelay := time.Duration(0)
	for index, device := range settings.Devices {
		placement := placements[device.Mac]
		if placement.Mac == "" {
			placement = defaultPlacement(device.Mac, index, len(settings.Devices))
		}
		delay := quantizeSceneDelay(time.Duration(math.Round(delayFactor(&sceneSettings, placement)*float64(stagger))) * time.Millisecond)
		if delay > maxDelay {
			maxDelay = delay
		}
		devices = append(devices, computedSceneDevice{
			mac:        device.Mac,
			delay:      delay,
			offset:     time.Duration(placement.TimingOffset) * time.Millisecond,
			transition: transitionForPreset(sceneSettings.Preset, duration, placement.Role),
		})
	}

	for _, device := range devices {
		target := sceneTargetForDevice(&sceneSettings, device.mac)
		baseDelay := device.offset
		delay := device.delay + device.offset
		if reverse {
			delay = maxDelay - device.delay + device.offset
		}

		if reverse {
			commands = append(commands, buildReverseColorTravelCommands(device.mac, delay, device.transition, target.Color, sceneSettings.ColorTravel)...)
			off := false
			powerOffDelay := delay + time.Duration(device.transition)*time.Millisecond
			commands = append(commands, sceneCommand{DeviceMac: device.mac, Delay: powerOffDelay, Power: &off})
			if target.Color.Brightness > 0 {
				commands = append(commands, sceneCommand{DeviceMac: device.mac, Delay: powerOffDelay + sceneOffRestoreDelay, Transition: 0, Color: &target.Color})
			}
			continue
		}
		if !target.PowerOn {
			off := false
			commands = append(commands, sceneCommand{DeviceMac: device.mac, Delay: delay, Transition: device.transition, Power: &off})
			continue
		}
		if sceneSettings.StartFromCurrent {
			on := true
			commands = append(commands, sceneCommand{DeviceMac: device.mac, Delay: delay, Transition: device.transition, Power: &on})
			if sceneSettings.ColorTravel != "none" {
				commands = append(commands, buildColorTravelCommands(device.mac, delay, device.transition, target.Color, sceneSettings.ColorTravel)...)
			} else {
				commands = append(commands, sceneCommand{DeviceMac: device.mac, Delay: delay, Transition: device.transition, Color: &target.Color})
			}
			continue
		}
		on := true
		start := target.Color
		start.Brightness = 0
		commands = append(commands,
			sceneCommand{DeviceMac: device.mac, Delay: baseDelay, Transition: 0, Color: &start},
			sceneCommand{DeviceMac: device.mac, Delay: baseDelay, Power: &on},
		)

		if sceneSettings.ColorTravel != "none" {
			commands = append(commands, buildColorTravelCommands(device.mac, delay, device.transition, target.Color, sceneSettings.ColorTravel)...)
		} else {
			commands = append(commands, sceneCommand{DeviceMac: device.mac, Delay: delay, Transition: device.transition, Color: &target.Color})
		}
	}

	commands = normalizeSceneCommandDelays(commands)
	sort.SliceStable(commands, func(i, j int) bool {
		return commands[i].Delay < commands[j].Delay
	})
	return commands
}

type computedSceneDevice struct {
	mac        string
	delay      time.Duration
	offset     time.Duration
	transition uint32
}

func normalizeSceneCommandDelays(commands []sceneCommand) []sceneCommand {
	minDelay := time.Duration(0)
	for _, command := range commands {
		if command.Delay < minDelay {
			minDelay = command.Delay
		}
	}
	if minDelay >= 0 {
		return commands
	}
	shift := -minDelay
	for index := range commands {
		commands[index].Delay += shift
	}
	return commands
}

func quantizeSceneDelay(delay time.Duration) time.Duration {
	if delay <= 0 {
		return 0
	}
	return time.Duration(math.Round(float64(delay)/float64(sceneDelayStep))) * sceneDelayStep
}

func normalizeSceneSettings(settings *models.SceneSettings) {
	if settings.Preset == "depart" {
		settings.Preset = defaultScenePreset
		settings.Reverse = true
	}
	if settings.Preset == "" || !isValidScenePreset(settings.Preset) {
		settings.Preset = defaultScenePreset
	}
	if settings.Direction == "" || !isValidSceneDirection(settings.Direction) {
		settings.Direction = defaultSceneDirection
	}
	if settings.ColorTravel == "" || !isValidSceneColorTravel(settings.ColorTravel) {
		settings.ColorTravel = defaultSceneColorTravel
	}
	if settings.Duration == 0 {
		settings.Duration = defaultSceneDuration
	}
	if settings.Duration > maxSceneDuration {
		settings.Duration = maxSceneDuration
	}
	if settings.Stagger > maxSceneStagger {
		settings.Stagger = maxSceneStagger
	}
	if settings.Intensity > 100 {
		settings.Intensity = 100
	}
	if len(settings.Devices) > maxSceneDevices {
		settings.Devices = settings.Devices[:maxSceneDevices]
	}
	if len(settings.Placements) > maxScenePlacements {
		settings.Placements = settings.Placements[:maxScenePlacements]
	}
	if len(settings.DeviceColors) > maxSceneDevices {
		settings.DeviceColors = settings.DeviceColors[:maxSceneDevices]
	}
	for index := range settings.Placements {
		settings.Placements[index].TimingOffset = int32(clampFloat(float64(settings.Placements[index].TimingOffset), -maxSceneTimingOffset, maxSceneTimingOffset))
	}
	settings.TargetColor = normalizeSceneColor(settings.TargetColor)
	for index := range settings.DeviceColors {
		settings.DeviceColors[index].Color = normalizeSceneColor(settings.DeviceColors[index].Color)
	}
}

func normalizeSceneColor(color models.Color) models.Color {
	color.Hue = uint(clampFloat(float64(color.Hue), 0, 360))
	color.Saturation = uint(clampFloat(float64(color.Saturation), 0, 100))
	color.Brightness = uint(clampFloat(float64(color.Brightness), 0, 100))
	if color.Kelvin == 0 {
		color.Kelvin = 3200
	}
	if color.Kelvin < 1500 {
		color.Kelvin = 1500
	}
	if color.Kelvin > 9000 {
		color.Kelvin = 9000
	}
	return color
}

func isValidScenePreset(preset string) bool {
	switch preset {
	case "arrival", "bloom", "sweep", "ignition", "breathe":
		return true
	}
	return false
}

func isValidSceneDirection(direction string) bool {
	switch direction {
	case "left-right", "right-left", "front-back", "back-front":
		return true
	}
	return false
}

func isValidSceneColorTravel(colorTravel string) bool {
	switch colorTravel {
	case "none", "warm", "hue":
		return true
	}
	return false
}

func buildColorTravelCommands(mac string, delay time.Duration, transition uint32, target golifx.HSBK, colorTravel string) []sceneCommand {
	switch colorTravel {
	case "warm":
		red := target
		red.Hue = uint16(0)
		red.Saturation = uint16(65500)
		red.Brightness = brightnessFloor(target.Brightness)
		amber := target
		amber.Hue = uint16(5460)
		amber.Saturation = uint16(58950)
		amber.Brightness = brightnessFloor(target.Brightness)
		return []sceneCommand{
			{DeviceMac: mac, Delay: delay, Transition: transition / 3, Color: &red},
			{DeviceMac: mac, Delay: delay + time.Duration(transition/3)*time.Millisecond, Transition: transition / 3, Color: &amber},
			{DeviceMac: mac, Delay: delay + time.Duration(2*(transition/3))*time.Millisecond, Transition: transition / 3, Color: &target},
		}
	case "hue":
		first := target
		first.Hue = shiftedHue(target.Hue, -60)
		first.Saturation = saturationFloor(target.Saturation)
		first.Brightness = brightnessFloor(target.Brightness)
		second := target
		second.Hue = shiftedHue(target.Hue, 45)
		second.Saturation = saturationFloor(target.Saturation)
		second.Brightness = brightnessFloor(target.Brightness)
		return []sceneCommand{
			{DeviceMac: mac, Delay: delay, Transition: transition / 3, Color: &first},
			{DeviceMac: mac, Delay: delay + time.Duration(transition/3)*time.Millisecond, Transition: transition / 3, Color: &second},
			{DeviceMac: mac, Delay: delay + time.Duration(2*(transition/3))*time.Millisecond, Transition: transition / 3, Color: &target},
		}
	}
	return []sceneCommand{{DeviceMac: mac, Delay: delay, Transition: transition, Color: &target}}
}

func buildReverseColorTravelCommands(mac string, delay time.Duration, transition uint32, target golifx.HSBK, colorTravel string) []sceneCommand {
	end := target
	end.Brightness = 0
	if colorTravel == "none" {
		return []sceneCommand{{DeviceMac: mac, Delay: delay, Transition: transition, Color: &end}}
	}

	segments := uint32(3)
	segmentTransition := transition / segments
	if segmentTransition == 0 {
		segmentTransition = transition
	}

	switch colorTravel {
	case "warm":
		red := target
		red.Hue = uint16(0)
		red.Saturation = uint16(65500)
		red.Brightness = brightnessFloor(target.Brightness)
		amber := target
		amber.Hue = uint16(5460)
		amber.Saturation = uint16(58950)
		amber.Brightness = brightnessFloor(target.Brightness)
		return []sceneCommand{
			{DeviceMac: mac, Delay: delay, Transition: segmentTransition, Color: &amber},
			{DeviceMac: mac, Delay: delay + time.Duration(segmentTransition)*time.Millisecond, Transition: segmentTransition, Color: &red},
			{DeviceMac: mac, Delay: delay + time.Duration(2*segmentTransition)*time.Millisecond, Transition: segmentTransition, Color: &end},
		}
	case "hue":
		first := target
		first.Hue = shiftedHue(target.Hue, -60)
		first.Saturation = saturationFloor(target.Saturation)
		first.Brightness = brightnessFloor(target.Brightness)
		second := target
		second.Hue = shiftedHue(target.Hue, 45)
		second.Saturation = saturationFloor(target.Saturation)
		second.Brightness = brightnessFloor(target.Brightness)
		return []sceneCommand{
			{DeviceMac: mac, Delay: delay, Transition: segmentTransition, Color: &second},
			{DeviceMac: mac, Delay: delay + time.Duration(segmentTransition)*time.Millisecond, Transition: segmentTransition, Color: &first},
			{DeviceMac: mac, Delay: delay + time.Duration(2*segmentTransition)*time.Millisecond, Transition: segmentTransition, Color: &end},
		}
	}
	return []sceneCommand{{DeviceMac: mac, Delay: delay, Transition: transition, Color: &end}}
}

func placementMap(settings *models.SceneSettings) map[string]models.SceneDevicePlacement {
	placements := make(map[string]models.SceneDevicePlacement)
	for _, placement := range settings.Placements {
		placement.X = clampFloat(placement.X, 0, 100)
		placement.Y = clampFloat(placement.Y, 0, 100)
		placement.TimingOffset = int32(clampFloat(float64(placement.TimingOffset), -maxSceneTimingOffset, maxSceneTimingOffset))
		placements[placement.Mac] = placement
	}
	return placements
}

func sceneTargetForDevice(settings *models.SceneSettings, mac string) sceneTarget {
	for _, deviceColor := range settings.DeviceColors {
		if deviceColor.Mac == mac {
			powerOn := true
			if deviceColor.PowerOn != nil {
				powerOn = *deviceColor.PowerOn
			} else if deviceColor.Color.PowerOn != nil {
				powerOn = *deviceColor.Color.PowerOn
			}
			return sceneTarget{Color: hsbkFromColor(deviceColor.Color), PowerOn: powerOn}
		}
	}
	return sceneTarget{Color: hsbkFromColor(settings.TargetColor), PowerOn: true}
}

func defaultPlacement(mac string, index int, total int) models.SceneDevicePlacement {
	if total <= 0 {
		total = 1
	}
	angle := (float64(index) / float64(total)) * 2 * math.Pi
	return models.SceneDevicePlacement{
		Mac: mac,
		X:   50 + math.Cos(angle)*35,
		Y:   50 + math.Sin(angle)*35,
	}
}

func delayFactor(settings *models.SceneSettings, placement models.SceneDevicePlacement) float64 {
	switch settings.Preset {
	case "arrival":
		switch placement.Role {
		case "overhead":
			return 0
		case "front", "side":
			return 0.4
		case "back":
			return 1
		default:
			return clampFloat(placement.Y/100, 0, 1)
		}
	case "bloom", "ignition", "breathe":
		x := placement.X - 50
		y := placement.Y - 50
		return clampFloat(math.Sqrt(x*x+y*y)/70.71, 0, 1)
	case "sweep":
		switch settings.Direction {
		case "right-left":
			return clampFloat((100-placement.X)/100, 0, 1)
		case "front-back":
			return clampFloat(placement.Y/100, 0, 1)
		case "back-front":
			return clampFloat((100-placement.Y)/100, 0, 1)
		default:
			return clampFloat(placement.X/100, 0, 1)
		}
	default:
		return 0
	}
}

func transitionForPreset(preset string, duration uint32, role string) uint32 {
	switch preset {
	case "arrival":
		switch role {
		case "overhead":
			return uint32(float64(duration) * 1.2)
		case "front", "side":
			return uint32(float64(duration) * 0.65)
		case "back":
			return duration
		}
	case "breathe":
		return uint32(float64(duration) * 1.15)
	}
	return duration
}

func hsbkFromColor(color models.Color) golifx.HSBK {
	hue, saturation, brightness, kelvin, _ := color.GenerateLIFXValues()
	return golifx.HSBK{
		Hue:        hue,
		Saturation: saturation,
		Brightness: brightness,
		Kelvin:     kelvin,
	}
}

func shiftedHue(hue uint16, degrees int) uint16 {
	value := int(hue) + degrees*182
	for value < 0 {
		value += 65535
	}
	return uint16(value % 65535)
}

func saturationFloor(target uint16) uint16 {
	minimum := uint16(32750)
	if target < minimum {
		return minimum
	}
	return target
}

func brightnessFloor(target uint16) uint16 {
	minimum := uint16(6550)
	if target < minimum {
		return target
	}
	return minimum
}

func clampFloat(value float64, min float64, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
