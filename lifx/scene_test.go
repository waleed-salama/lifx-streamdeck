package lifx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/wwsean08/lifx-streamdeck/models"
)

func TestBuildSceneCommandsArrivalUsesRoles(t *testing.T) {
	settings := &models.SceneSettings{
		Preset:    "arrival",
		Duration:  1000,
		Stagger:   1000,
		Intensity: 50,
		Devices: []models.Device{
			{Name: "Overhead", Mac: "d0:73:d5:2b:a7:b8"},
			{Name: "Front", Mac: "d0:73:d5:3c:86:05"},
			{Name: "Back", Mac: "d0:73:d5:3c:dc:6d"},
		},
		Placements: []models.SceneDevicePlacement{
			{Mac: "d0:73:d5:2b:a7:b8", X: 50, Y: 10, Role: "overhead"},
			{Mac: "d0:73:d5:3c:86:05", X: 15, Y: 25, Role: "front"},
			{Mac: "d0:73:d5:3c:dc:6d", X: 50, Y: 90, Role: "back"},
		},
		TargetColor: models.Color{Hue: 30, Saturation: 0, Brightness: 80, Kelvin: 3200},
	}

	commands := buildSceneCommands(settings)

	require.NotEmpty(t, commands)
	require.Contains(t, colorDelays(commands, "d0:73:d5:2b:a7:b8"), time.Duration(0))
	require.Contains(t, colorDelays(commands, "d0:73:d5:3c:86:05"), 400*time.Millisecond)
	require.Contains(t, colorDelays(commands, "d0:73:d5:3c:dc:6d"), time.Second)
	require.Equal(t, uint32(1200), transitionForDevice(commands, "d0:73:d5:2b:a7:b8", 0))
	require.Equal(t, uint32(650), transitionForDevice(commands, "d0:73:d5:3c:86:05", 400*time.Millisecond))
}

func TestBuildSceneCommandsPreloadsDarkColorBeforePowerOn(t *testing.T) {
	settings := &models.SceneSettings{
		Preset:      "sweep",
		Duration:    1000,
		Devices:     []models.Device{{Name: "Desk", Mac: "d0:73:d5:2b:a7:b8"}},
		TargetColor: models.Color{Hue: 30, Saturation: 0, Brightness: 80, Kelvin: 3200},
	}

	commands := buildSceneCommands(settings)

	require.NotNil(t, commands[0].Color)
	require.Zero(t, commands[0].Color.Brightness)
	require.Nil(t, commands[0].Power)
	require.NotNil(t, commands[1].Power)
	require.True(t, *commands[1].Power)
}

func TestBuildSceneCommandsQuantizesNearbySweepDelays(t *testing.T) {
	settings := &models.SceneSettings{
		Preset:    "sweep",
		Direction: "front-back",
		Duration:  1000,
		Stagger:   2000,
		Intensity: 70,
		Devices: []models.Device{
			{Name: "Left", Mac: "d0:73:d5:01:df:8f"},
			{Name: "Right", Mac: "d0:73:d5:01:be:91"},
		},
		Placements: []models.SceneDevicePlacement{
			{Mac: "d0:73:d5:01:df:8f", X: 33.7, Y: 8.38},
			{Mac: "d0:73:d5:01:be:91", X: 66.4, Y: 7.35},
		},
		TargetColor: models.Color{Hue: 30, Saturation: 0, Brightness: 80, Kelvin: 3200},
	}

	commands := buildSceneCommands(settings)

	require.Equal(t, colorDelays(commands, "d0:73:d5:01:df:8f"), colorDelays(commands, "d0:73:d5:01:be:91"))
}

func TestGroupSceneCommandsByDelay(t *testing.T) {
	commands := []sceneCommand{
		{DeviceMac: "a", Delay: 0},
		{DeviceMac: "b", Delay: 0},
		{DeviceMac: "a", Delay: 50 * time.Millisecond},
	}

	groups := groupSceneCommandsByDelay(commands)

	require.Len(t, groups, 2)
	require.Len(t, groups[0], 2)
	require.Len(t, groups[1], 1)
}

func TestBuildSceneCommandsIgnitionAddsColorTravel(t *testing.T) {
	settings := &models.SceneSettings{
		Preset:      "ignition",
		ColorTravel: "warm",
		Duration:    900,
		Stagger:     0,
		Devices:     []models.Device{{Name: "Desk", Mac: "d0:73:d5:2b:a7:b8"}},
		Placements: []models.SceneDevicePlacement{
			{Mac: "d0:73:d5:2b:a7:b8", X: 50, Y: 50},
		},
		TargetColor: models.Color{
			Hue:        30,
			Saturation: 0,
			Brightness: 80,
			Kelvin:     3200,
		},
	}

	commands := buildSceneCommands(settings)
	colors := 0
	for _, command := range commands {
		if command.Color != nil {
			colors++
		}
	}

	require.Equal(t, 4, colors)
	require.Contains(t, colorDelays(commands, "d0:73:d5:2b:a7:b8"), 0*time.Millisecond)
	require.Contains(t, colorDelays(commands, "d0:73:d5:2b:a7:b8"), 300*time.Millisecond)
	require.Contains(t, colorDelays(commands, "d0:73:d5:2b:a7:b8"), 600*time.Millisecond)
}

func TestBuildSceneCommandsHueTravelUsesHueShift(t *testing.T) {
	settings := &models.SceneSettings{
		Preset:      "arrival",
		ColorTravel: "hue",
		Duration:    900,
		Devices:     []models.Device{{Name: "Desk", Mac: "d0:73:d5:2b:a7:b8"}},
		TargetColor: models.Color{
			Hue:        120,
			Saturation: 80,
			Brightness: 80,
			Kelvin:     3200,
		},
	}

	commands := buildSceneCommands(settings)

	require.NotEqual(t, uint16(0), commands[2].Color.Hue)
	require.NotEqual(t, uint16(5460), commands[3].Color.Hue)
	require.Equal(t, uint16(21840), commands[4].Color.Hue)
}

func TestNormalizeSceneSettingsClampsUntrustedPayload(t *testing.T) {
	settings := &models.SceneSettings{
		Preset:      "bad",
		Direction:   "bad",
		ColorTravel: "bad",
		Duration:    90000,
		Stagger:     90000,
		Intensity:   200,
		TargetColor: models.Color{Hue: 500, Saturation: 500, Brightness: 500, Kelvin: 20000},
	}
	for i := 0; i < maxSceneDevices+5; i++ {
		settings.Devices = append(settings.Devices, models.Device{Mac: "d0:73:d5:2b:a7:b8"})
	}
	settings.Placements = []models.SceneDevicePlacement{{Mac: "d0:73:d5:2b:a7:b8", TimingOffset: -90000}}

	normalizeSceneSettings(settings)

	require.Equal(t, defaultScenePreset, settings.Preset)
	require.Equal(t, defaultSceneDirection, settings.Direction)
	require.Equal(t, defaultSceneColorTravel, settings.ColorTravel)
	require.Equal(t, uint32(maxSceneDuration), settings.Duration)
	require.Equal(t, uint32(maxSceneStagger), settings.Stagger)
	require.Equal(t, uint(100), settings.Intensity)
	require.Len(t, settings.Devices, maxSceneDevices)
	require.Equal(t, int32(-maxSceneTimingOffset), settings.Placements[0].TimingOffset)
	require.Equal(t, uint(360), settings.TargetColor.Hue)
	require.Equal(t, uint(100), settings.TargetColor.Saturation)
	require.Equal(t, uint(100), settings.TargetColor.Brightness)
	require.Equal(t, uint16(9000), settings.TargetColor.Kelvin)
}

func TestBuildSceneCommandsReverseArrivalUsesMirroredRoleOrder(t *testing.T) {
	settings := &models.SceneSettings{
		Preset:    "arrival",
		Reverse:   true,
		Duration:  1000,
		Stagger:   1000,
		Intensity: 50,
		Devices: []models.Device{
			{Name: "Overhead", Mac: "d0:73:d5:2b:a7:b8"},
			{Name: "Front", Mac: "d0:73:d5:3c:86:05"},
			{Name: "Back", Mac: "d0:73:d5:3c:dc:6d"},
		},
		Placements: []models.SceneDevicePlacement{
			{Mac: "d0:73:d5:2b:a7:b8", X: 50, Y: 10, Role: "overhead"},
			{Mac: "d0:73:d5:3c:86:05", X: 15, Y: 25, Role: "front"},
			{Mac: "d0:73:d5:3c:dc:6d", X: 50, Y: 90, Role: "back"},
		},
		TargetColor: models.Color{Hue: 30, Saturation: 0, Brightness: 80, Kelvin: 3200},
	}

	commands := buildSceneCommands(settings)

	require.Contains(t, darkColorDelays(commands, "d0:73:d5:3c:dc:6d"), 0*time.Millisecond)
	require.Contains(t, darkColorDelays(commands, "d0:73:d5:3c:86:05"), 600*time.Millisecond)
	require.Contains(t, darkColorDelays(commands, "d0:73:d5:2b:a7:b8"), time.Second)
	require.Contains(t, powerOffDelays(commands, "d0:73:d5:3c:dc:6d"), time.Second)
	require.Contains(t, powerOffDelays(commands, "d0:73:d5:3c:86:05"), 1250*time.Millisecond)
	require.Contains(t, powerOffDelays(commands, "d0:73:d5:2b:a7:b8"), 2200*time.Millisecond)
}

func TestBuildSceneCommandsReverseWarmTravelWalksBackThroughColor(t *testing.T) {
	settings := &models.SceneSettings{
		Preset:      "ignition",
		Reverse:     true,
		ColorTravel: "warm",
		Duration:    900,
		Devices:     []models.Device{{Name: "Desk", Mac: "d0:73:d5:2b:a7:b8"}},
		TargetColor: models.Color{Hue: 30, Saturation: 0, Brightness: 80, Kelvin: 3200},
	}

	commands := buildSceneCommands(settings)

	require.Len(t, commands, 5)
	require.Equal(t, uint16(5460), commands[0].Color.Hue)
	require.Greater(t, commands[0].Color.Brightness, uint16(0))
	require.Equal(t, uint16(0), commands[1].Color.Hue)
	require.Greater(t, commands[1].Color.Brightness, uint16(0))
	require.Zero(t, commands[2].Color.Brightness)
	require.NotNil(t, commands[3].Power)
	require.False(t, *commands[3].Power)
	require.Equal(t, 900*time.Millisecond, commands[3].Delay)
	require.NotNil(t, commands[4].Color)
	require.Equal(t, commands[3].Delay+sceneOffRestoreDelay, commands[4].Delay)
	require.Greater(t, commands[4].Color.Brightness, uint16(0))
}

func TestBuildSceneCommandsTimingOffsetCanLeadBeyondSceneStart(t *testing.T) {
	settings := &models.SceneSettings{
		Preset:    "sweep",
		Reverse:   true,
		Direction: "front-back",
		Duration:  2000,
		Stagger:   1000,
		Intensity: 50,
		Devices: []models.Device{
			{Name: "Front", Mac: "d0:73:d5:3c:86:05"},
			{Name: "Back", Mac: "d0:73:d5:3c:dc:6d"},
		},
		Placements: []models.SceneDevicePlacement{
			{Mac: "d0:73:d5:3c:86:05", X: 50, Y: 0},
			{Mac: "d0:73:d5:3c:dc:6d", X: 50, Y: 100, TimingOffset: -3000},
		},
		TargetColor: models.Color{Hue: 30, Saturation: 0, Brightness: 80, Kelvin: 3200},
	}

	commands := buildSceneCommands(settings)

	require.Contains(t, darkColorDelays(commands, "d0:73:d5:3c:dc:6d"), 0*time.Millisecond)
	require.Contains(t, darkColorDelays(commands, "d0:73:d5:3c:86:05"), 4*time.Second)
	require.Contains(t, powerOffDelays(commands, "d0:73:d5:3c:dc:6d"), 2*time.Second)
	require.Contains(t, powerOffDelays(commands, "d0:73:d5:3c:86:05"), 6*time.Second)
}

func TestNormalizeSceneSettingsMigratesDepartToReverseArrival(t *testing.T) {
	settings := &models.SceneSettings{Preset: "depart"}

	normalizeSceneSettings(settings)

	require.Equal(t, defaultScenePreset, settings.Preset)
	require.True(t, settings.Reverse)
}

func colorDelays(commands []sceneCommand, mac string) []time.Duration {
	delays := make([]time.Duration, 0)
	for _, command := range commands {
		if command.DeviceMac == mac && command.Color != nil && command.Color.Brightness > 0 {
			delays = append(delays, command.Delay)
		}
	}
	return delays
}

func darkColorDelays(commands []sceneCommand, mac string) []time.Duration {
	delays := make([]time.Duration, 0)
	for _, command := range commands {
		if command.DeviceMac == mac && command.Color != nil && command.Color.Brightness == 0 {
			delays = append(delays, command.Delay)
		}
	}
	return delays
}

func powerOffDelays(commands []sceneCommand, mac string) []time.Duration {
	delays := make([]time.Duration, 0)
	for _, command := range commands {
		if command.DeviceMac == mac && command.Power != nil && !*command.Power {
			delays = append(delays, command.Delay)
		}
	}
	return delays
}

func transitionForDevice(commands []sceneCommand, mac string, delay time.Duration) uint32 {
	for _, command := range commands {
		if command.DeviceMac == mac && command.Delay == delay && command.Color != nil && command.Color.Brightness > 0 {
			return command.Transition
		}
	}
	return 0
}
