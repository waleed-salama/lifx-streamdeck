package lifx

import (
	"github.com/stretchr/testify/require"
	"gitlab.com/wwsean08/golifx"
	"gitlab.com/wwsean08/lifx-streamdeck/models"
	"sync"
	"testing"
)

func TestNewLifxController(t *testing.T) {
	client := NewLifxController()
	require.IsType(t, &lifxClient{}, client)
	specificClient := client.(*lifxClient)

	require.NotNil(t, specificClient.deviceMap)
	require.NotNil(t, specificClient.writeLock)
	require.NotNil(t, specificClient.sceneLock)

}

func TestLifxClient_clearDeviceMap(t *testing.T) {
	client := new(lifxClient)
	client.writeLock = new(sync.Mutex)
	require.Nil(t, client.deviceMap)
	client.addDeviceToMap([]*golifx.Device{new(golifx.Device)})
	require.Len(t, client.deviceMap, 1)
	client.clearDeviceMap()
	require.Len(t, client.deviceMap, 0)
}

func TestLifxClient_addDevicesToMap(t *testing.T) {
	client := new(lifxClient)
	client.writeLock = new(sync.Mutex)
	require.Nil(t, client.deviceMap)
	device := new(golifx.Device)
	device.SetHardwareAddress(203028134458320)
	client.addDeviceToMap([]*golifx.Device{device})

	devices := client.GetDevices()
	require.Len(t, devices, 1)
	require.Equal(t, device, devices["d0:73:d5:2b:a7:b8"])

	// try a second time when it's not been cleared
	client.clearDeviceMap()
	client.addDeviceToMap([]*golifx.Device{device})

	devices = client.GetDevices()
	require.Len(t, devices, 1)
	require.Equal(t, device, devices["d0:73:d5:2b:a7:b8"])
}

func TestMacToUint64(t *testing.T) {
	mac, err := macToUint64("d0:73:d5:2b:a7:b8")
	require.NoError(t, err)
	require.Equal(t, uint64(0xb8a72bd573d00000), mac)

	mac, err = macToUint64("z")
	require.Equal(t, uint64(0), mac)
	require.Error(t, err)
}

func TestDeviceFromMac(t *testing.T) {
	device, err := deviceFromMac("d0:73:d5:2b:a7:b8")
	require.NoError(t, err)
	require.NotNil(t, device)
	require.Equal(t, "d0:73:d5:2b:a7:b8", device.MacAddress())

	device, err = deviceFromMac("z")
	require.Error(t, err)
	require.Nil(t, device)
}

func TestResolveDevicesSkipsInvalidMacEntries(t *testing.T) {
	client := &lifxClient{
		deviceMap: make(map[string]*golifx.Device),
		writeLock: new(sync.Mutex),
	}
	device, err := deviceFromMac("d0:73:d5:2b:a7:b8")
	require.NoError(t, err)
	client.addDeviceToMap([]*golifx.Device{device})

	devices, err := client.resolveDevices([]models.Device{
		{Name: "Beam", Mac: "d0:73:d5:2b:a7:b8"},
		{Name: "Bad", Mac: "Top"},
	})

	require.NoError(t, err)
	require.Len(t, devices, 1)
	require.Equal(t, device, devices["d0:73:d5:2b:a7:b8"])
}

func TestResolveDevicesReturnsErrorWhenNoValidMacsSelected(t *testing.T) {
	client := &lifxClient{
		deviceMap: make(map[string]*golifx.Device),
		writeLock: new(sync.Mutex),
	}

	devices, err := client.resolveDevices([]models.Device{{Name: "Bad", Mac: "Top"}})

	require.Error(t, err)
	require.Nil(t, devices)
}

func TestBroadcastOnlyCache(t *testing.T) {
	client := &lifxClient{writeLock: new(sync.Mutex)}
	mac := "d0:73:d5:10:97:74"

	require.False(t, client.shouldUseBroadcastOnly(mac))
	client.markBroadcastOnly(mac)
	require.True(t, client.shouldUseBroadcastOnly(mac))
	client.clearBroadcastOnly(mac)
	require.False(t, client.shouldUseBroadcastOnly(mac))
}

func TestLifxClient_UpdateGlobalSettings(t *testing.T) {
	client := new(lifxClient)
	client.writeLock = new(sync.Mutex)
	gsettings := new(models.GlobalSettings)
	gsettings.OutIP = "1.2.3.4"
	gsettings.DirectComm = true

	err := client.UpdateGlobalSettings(gsettings)
	require.NoError(t, err)
	require.Equal(t, "1.2.3.4", client.gSettings.OutIP)
	require.True(t, client.gSettings.DirectComm)
}

func TestLifxClient_startSceneCancelsPreviousScene(t *testing.T) {
	client := new(lifxClient)
	first := client.startScene()
	second := client.startScene()

	select {
	case <-first:
	default:
		t.Fatal("expected previous scene to be canceled")
	}
	select {
	case <-second:
		t.Fatal("new scene should stay active")
	default:
	}
}

func TestLifxClient_cancelSceneCancelsActiveScene(t *testing.T) {
	client := new(lifxClient)
	active := client.startScene()

	client.cancelScene()

	select {
	case <-active:
	default:
		t.Fatal("expected active scene to be canceled")
	}
}
