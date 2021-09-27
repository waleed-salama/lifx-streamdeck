package models

type (
	GlobalSettings struct {
		OutIP         string   `mapstructure:"outIP" json:"out_ip"`
		Version       int      `mapstructure:"version" json:"version"`
		DirectComm    bool     `mapstructure:"directComm" json:"directComm"`
		CustomDevices []string `mapstructure:"customDevices" json:"customDevices"`
	}

	//Device represents a device in the Stream Deck settings
	Device struct {
		Name string `mapstructure:"name"`
		Mac  string `mapstructure:"mac"`
	}

	//Color represents the color in settings
	Color struct {
		Hue        uint   `mapstructure:"hue"`
		Saturation uint   `mapstructure:"saturation"`
		Brightness uint   `mapstructure:"brightness"`
		Kelvin     uint16 `mapstructure:"kelvin"`
		Transition uint32 `mapstructure:"transition"`
	}

	//PowerSettings represents the on/off action settings
	PowerSettings struct {
		Version    int      `mapstructure:"version"`
		Transition uint32   `mapstructure:"transition"`
		Devices    []Device `mapstructure:"devices"`
	}

	ToggleSettings struct {
		Version    int      `mapstructure:"version"`
		Transition uint32   `mapstructure:"transition"`
		Devices    []Device `mapstructure:"devices"`
	}

	//ColorSettings represents the set color action settings
	ColorSettings struct {
		Version int      `mapstructure:"version"`
		Color   Color    `mapstructure:"color"`
		Devices []Device `mapstructure:"devices"`
	}

	//BrightnessSettings represents the brightness settings
	BrightnessSettings struct {
		Version    int      `mapstructure:"version"`
		Devices    []Device `mapstructure:"devices"`
		Brightness uint     `mapstructure:"brightness"`
		Transition uint32   `mapstructure:"transition"`
	}

	//WaveFormSettings represents the waveform settings
	WaveFormSettings struct {
		Version        int      `mapstructure:"version"`
		IsTransient    bool     `mapstructure:"isTransient"`
		Period         uint32   `mapstructure:"period"`
		Cycles         float32  `mapstructure:"cycles"`
		Skew           int16    `mapstructure:"skew"`
		Waveform       uint8    `mapstructure:"waveform"`
		Devices        []Device `mapstructure:"devices"`
		AlternateColor Color    `mapstructure:"color"`
	}

	Debug struct {
		NetInfo    []byte `mapstructure:"netInfo" json:"netInfo"`
		DeviceData []byte `mapstructure:"deviceData" json:"deviceData"`
	}

	MultizoneSettings struct {
		Devices    []Device `mapstructure:"device"`
		Colors     []Color  `mapstructure:"colors"`
		Transition uint32   `mapstructure:"transition"`
	}
)

// Application used for debug
type Application struct {
	Language string `json:"language"`
	Platform string `json:"platform"`
	Version  string `json:"version"`
}

// Plugin used for debug
type Plugin struct {
	Version string `json:"version"`
}

type AppInfo struct {
	Application Application `json:"application"`
	Plugin      Plugin      `json:"plugin"`
	Version     string      `json:"version"`
	Commit      string      `json:"commit"`
}

//GenerateLIFXValues is a helper function to take care of
//the math convertting user input to proper values for lifx
func (c Color) GenerateLIFXValues() (uint16, uint16, uint16, uint16, uint32) {
	hue := 182 * c.Hue
	sat := 655 * c.Saturation
	bri := 655 * c.Brightness

	return uint16(hue), uint16(sat), uint16(bri), c.Kelvin, c.Transition
}

//GetSkewRatio converts the input skew into a format that LIFX
// understands
func (w WaveFormSettings) GetSkewRatio() int16 {
	return w.Skew * 327
}
