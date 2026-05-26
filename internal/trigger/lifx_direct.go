package trigger

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	"gitlab.com/wwsean08/lifx-streamdeck/models"
)

const (
	lifxPort            = 56700
	lifxSetWaveformType = 103
	lifxHeaderLength    = 36
	lifxDefaultSource   = 7
)

func SendWaveformDirect(settings *models.WaveFormSettings, outIP string, deviceIPs map[string]string, attempts int, delay time.Duration) error {
	if settings == nil {
		return fmt.Errorf("waveform settings are required")
	}
	if attempts < 1 {
		attempts = 1
	}
	localAddr := &net.UDPAddr{Port: 0}
	if strings.TrimSpace(outIP) != "" {
		localAddr.IP = net.ParseIP(strings.TrimSpace(outIP))
		if localAddr.IP == nil {
			return fmt.Errorf("invalid outbound IP %q", outIP)
		}
	}
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	type targetPacket struct {
		name   string
		addr   *net.UDPAddr
		packet []byte
	}

	targets := make([]targetPacket, 0, len(settings.Devices))
	for _, device := range settings.Devices {
		ip := strings.TrimSpace(deviceIPs[normalizeMAC(device.Mac)])
		if ip == "" {
			return fmt.Errorf("missing unicast IP for %s", deviceName(device))
		}
		targetIP := net.ParseIP(ip)
		if targetIP == nil {
			return fmt.Errorf("invalid unicast IP %q for %s", ip, deviceName(device))
		}
		packet, err := buildWaveformPacket(settings, device.Mac)
		if err != nil {
			return err
		}
		targets = append(targets, targetPacket{
			name:   deviceName(device),
			addr:   &net.UDPAddr{IP: targetIP, Port: lifxPort},
			packet: packet,
		})
	}

	for attempt := 0; attempt < attempts; attempt++ {
		for _, target := range targets {
			if _, err := conn.WriteToUDP(target.packet, target.addr); err != nil {
				return err
			}
		}
		if attempt+1 < attempts {
			time.Sleep(delay)
		}
	}
	return nil
}

func buildWaveformPacket(settings *models.WaveFormSettings, mac string) ([]byte, error) {
	target, err := macTarget(mac)
	if err != nil {
		return nil, err
	}
	hue, saturation, brightness, kelvin, _ := settings.AlternateColor.GenerateLIFXValues()
	payload := make([]byte, 21)
	payload[1] = boolByte(settings.IsTransient)
	binary.LittleEndian.PutUint16(payload[2:4], hue)
	binary.LittleEndian.PutUint16(payload[4:6], saturation)
	binary.LittleEndian.PutUint16(payload[6:8], brightness)
	binary.LittleEndian.PutUint16(payload[8:10], kelvin)
	binary.LittleEndian.PutUint32(payload[10:14], settings.Period)
	binary.LittleEndian.PutUint32(payload[14:18], math.Float32bits(settings.Cycles))
	binary.LittleEndian.PutUint16(payload[18:20], uint16(settings.GetSkewRatio()))
	payload[20] = settings.Waveform

	packet := make([]byte, lifxHeaderLength+len(payload))
	binary.LittleEndian.PutUint16(packet[0:2], uint16(len(packet)))
	packet[3] = (1 << 4) | (1 << 2)
	binary.LittleEndian.PutUint32(packet[4:8], lifxDefaultSource)
	binary.LittleEndian.PutUint64(packet[8:16], target)
	packet[22] = 1
	binary.LittleEndian.PutUint16(packet[32:34], lifxSetWaveformType)
	copy(packet[lifxHeaderLength:], payload)
	return packet, nil
}

func macTarget(mac string) (uint64, error) {
	clean := strings.ReplaceAll(mac, ":", "")
	if len(clean) != 12 {
		return 0, fmt.Errorf("invalid MAC address %q", mac)
	}
	var bytes [8]byte
	for i := 0; i < 6; i++ {
		var value byte
		_, err := fmt.Sscanf(clean[i*2:i*2+2], "%02x", &value)
		if err != nil {
			return 0, fmt.Errorf("invalid MAC address %q", mac)
		}
		bytes[i] = value
	}
	return binary.LittleEndian.Uint64(bytes[:]), nil
}

func normalizeMAC(mac string) string {
	return strings.ToLower(strings.TrimSpace(mac))
}

func boolByte(value bool) byte {
	if value {
		return 1
	}
	return 0
}
