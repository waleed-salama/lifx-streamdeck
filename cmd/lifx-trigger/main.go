package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"gitlab.com/wwsean08/golifx"
	"gitlab.com/wwsean08/lifx-streamdeck/internal/trigger"
	"gitlab.com/wwsean08/lifx-streamdeck/lifx"
	"gitlab.com/wwsean08/lifx-streamdeck/models"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] == "waveform" {
		if len(args) > 0 {
			args = args[1:]
		}
		return runWaveform(args)
	}
	return fmt.Errorf("unknown command %q; expected waveform", args[0])
}

func runWaveform(args []string) error {
	streamDeckDir, err := trigger.DefaultStreamDeckDir()
	if err != nil {
		return err
	}

	flags := flag.NewFlagSet("waveform", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	title := flags.String("title", "Codex Done", "Stream Deck key title to replay")
	settingsPath := flags.String("settings", "", "waveform settings JSON file")
	flags.StringVar(&streamDeckDir, "streamdeck-dir", streamDeckDir, "Stream Deck application support directory")
	direct := flags.Bool("direct", true, "send commands by direct unicast to resolved device IPs")
	outIP := flags.String("out-ip", "", "local outbound IP to use for LIFX traffic")
	customDeviceIPs := newStringListFlag()
	flags.Var(customDeviceIPs, "custom-device-ip", "custom LIFX device IP to resolve before sending; repeat for multiple bulbs")
	deviceIPs := newDeviceIPFlag()
	flags.Var(deviceIPs, "device-ip", "direct unicast target in mac=ip form; repeat for multiple bulbs")
	includeMACs := newStringListFlag()
	flags.Var(includeMACs, "include-mac", "only trigger this selected MAC; repeat for multiple bulbs")
	attempts := flags.Int("attempts", 1, "direct unicast packet sends per selected bulb")
	wait := flags.Duration("wait", 2*time.Second, "time to keep the process alive after sending the waveform")
	verbose := flags.Bool("verbose", false, "print LIFX command progress")
	if err := flags.Parse(args); err != nil {
		return err
	}

	settings, source, err := resolveWaveformSettings(*settingsPath, streamDeckDir, *title)
	if err != nil {
		return err
	}
	settings.Devices = filterDevices(settings.Devices, includeMACs.values)
	if len(settings.Devices) == 0 {
		return fmt.Errorf("no waveform devices remain after include filter")
	}

	if len(deviceIPs.values) > 0 {
		if *verbose {
			fmt.Fprintf(os.Stderr, "triggering direct unicast waveform from %s via %s for %s\n", source, strings.TrimSpace(*outIP), deviceNames(settings.Devices))
		}
		if err := trigger.SendWaveformDirect(settings, strings.TrimSpace(*outIP), deviceIPs.values, *attempts, 75*time.Millisecond); err != nil {
			return err
		}
		time.Sleep(*wait)
		return nil
	}

	controller := lifx.NewLifxController()
	if logger, ok := controller.(interface{ SetLogger(func(string)) }); ok && *verbose {
		logger.SetLogger(func(message string) {
			fmt.Fprintln(os.Stderr, message)
		})
	}
	if err := controller.UpdateGlobalSettings(&models.GlobalSettings{
		Version:       4,
		DirectComm:    *direct,
		OutIP:         strings.TrimSpace(*outIP),
		CustomDevices: customDeviceIPs.values,
	}); err != nil {
		return err
	}
	if err := controller.DiscoverDevices(); err != nil {
		return err
	}
	if *direct {
		if err := ensureSelectedDevicesResolved(controller.GetDevices(), settings.Devices); err != nil {
			return err
		}
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "triggering waveform from %s for %s\n", source, deviceNames(settings.Devices))
	}
	if err := controller.SetWaveform(settings); err != nil {
		return err
	}
	time.Sleep(*wait)
	return nil
}

func resolveWaveformSettings(settingsPath, streamDeckDir, title string) (*models.WaveFormSettings, string, error) {
	if strings.TrimSpace(settingsPath) != "" {
		settings, err := trigger.LoadWaveformSettings(settingsPath)
		return settings, settingsPath, err
	}
	settings, path, err := trigger.FindWaveformSettingsByTitle(streamDeckDir, title)
	return settings, path, err
}

func deviceNames(devices []models.Device) string {
	names := make([]string, 0, len(devices))
	for _, device := range devices {
		if device.Name != "" {
			names = append(names, device.Name)
			continue
		}
		names = append(names, device.Mac)
	}
	return strings.Join(names, ", ")
}

func ensureSelectedDevicesResolved(discovered map[string]*golifx.Device, selected []models.Device) error {
	missing := make([]string, 0)
	for _, device := range selected {
		if discovered[device.Mac] == nil {
			missing = append(missing, deviceName(device))
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("direct mode could not resolve selected device(s): %s; add --custom-device-ip for those bulbs or run with --direct=false to allow broadcast fallback", strings.Join(missing, ", "))
	}
	return nil
}

func deviceName(device models.Device) string {
	return trigger.DeviceName(device)
}

func filterDevices(devices []models.Device, includeMACs []string) []models.Device {
	if len(includeMACs) == 0 {
		return devices
	}
	included := make(map[string]bool, len(includeMACs))
	for _, mac := range includeMACs {
		included[strings.ToLower(strings.TrimSpace(mac))] = true
	}
	filtered := make([]models.Device, 0, len(devices))
	for _, device := range devices {
		if included[strings.ToLower(strings.TrimSpace(device.Mac))] {
			filtered = append(filtered, device)
		}
	}
	return filtered
}

type stringListFlag struct {
	values []string
}

func newStringListFlag() *stringListFlag {
	return &stringListFlag{values: []string{}}
}

func (f *stringListFlag) String() string {
	return strings.Join(f.values, ",")
}

func (f *stringListFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	f.values = append(f.values, value)
	return nil
}

type deviceIPFlag struct {
	values map[string]string
}

func newDeviceIPFlag() *deviceIPFlag {
	return &deviceIPFlag{values: map[string]string{}}
}

func (f *deviceIPFlag) String() string {
	values := make([]string, 0, len(f.values))
	for mac, ip := range f.values {
		values = append(values, mac+"="+ip)
	}
	return strings.Join(values, ",")
}

func (f *deviceIPFlag) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("device-ip must be in mac=ip form")
	}
	mac := strings.ToLower(strings.TrimSpace(parts[0]))
	ip := strings.TrimSpace(parts[1])
	if mac == "" || ip == "" {
		return fmt.Errorf("device-ip must include both mac and ip")
	}
	f.values[mac] = ip
	return nil
}
