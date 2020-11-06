package lifx

type (
	//Device represents a device in the Stream Deck settings
	Device struct {
		Name string `mapstructure:"name"`
		Mac  string `mapstructure:"mac"`
	}

	//Color represents the color in settings
	Color struct {
		Hue        uint `mapstructure:"hue"`
		Saturation uint `mapstructure:"saturation"`
		Brightness uint `mapstructure:"brightness"`
		Kelvin     uint `mapstructure:"kelvin"`
		Transition uint `mapstructure:"transition"`
	}

	//PowerSettings represents the on/off action settings
	PowerSettings struct {
		Version int      `mapstructure:"version"`
		Devices []Device `mapstructure:"devices"`
	}

	ToggleSettings struct {
		Version int      `mapstructure:"version"`
		Devices []Device `mapstructure:"devices"`
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
		Skew           int      `mapstructure:"skew"`
		Waveform       uint8    `mapstructure:"waveform"`
		Devices        []Device `mapstructure:"devices"`
		AlternateColor Color    `mapstructure:"color"`
	}
)

//GenerateLIFXValues is a helper function to take care of
//the math convertting user input to proper values for lifx
func (c Color) GenerateLIFXValues() (uint16, uint16, uint16, uint16, uint32) {
	hue := 182 * c.Hue
	sat := 655 * c.Saturation
	bri := 655 * c.Brightness

	return uint16(hue), uint16(sat), uint16(bri), uint16(c.Kelvin), uint32(c.Transition)
}

//GetSkewRatio converts the input skew into a format that LIFX
// understands
func (w WaveFormSettings) GetSkewRatio() int16 {
	return int16(w.Skew * 327)
}
