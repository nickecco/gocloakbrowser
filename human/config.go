package human

import "time"

// Preset identifies human behavioral presets.
type Preset string

const (
	PresetDefault Preset = "default"
	PresetCareful Preset = "careful"
)

// Config controls parameters for human-like behavior emulation.
type Config struct {
	// Mouse movement
	MouseTimeBase      time.Duration
	MouseTimePerBit    time.Duration
	MouseTimeSpread    float64
	MouseCurvature     float64
	MouseOvershootProb float64
	MouseJitterProb    float64

	// Keyboard
	KeyDelayBase   time.Duration
	KeyDelaySpread float64

	// Scroll
	ScrollStepBase int
	ScrollDelay    time.Duration
}

// DefaultConfig returns default humanize parameters.
func DefaultConfig() *Config {
	return &Config{
		MouseTimeBase:      120 * time.Millisecond,
		MouseTimePerBit:    40 * time.Millisecond,
		MouseTimeSpread:    0.25,
		MouseCurvature:     0.3,
		MouseOvershootProb: 0.25,
		MouseJitterProb:    0.15,

		KeyDelayBase:   55 * time.Millisecond,
		KeyDelaySpread: 0.35,

		ScrollStepBase: 120,
		ScrollDelay:    20 * time.Millisecond,
	}
}

// CarefulConfig returns slower, more deliberate humanize parameters.
func CarefulConfig() *Config {
	cfg := DefaultConfig()
	cfg.MouseTimeBase = 200 * time.Millisecond
	cfg.MouseTimePerBit = 60 * time.Millisecond
	cfg.MouseCurvature = 0.4
	cfg.MouseOvershootProb = 0.15
	cfg.KeyDelayBase = 95 * time.Millisecond
	cfg.KeyDelaySpread = 0.45
	cfg.ScrollDelay = 35 * time.Millisecond
	return cfg
}

// ResolveConfig returns the Config based on preset and optional overrides.
func ResolveConfig(preset Preset, custom *Config) *Config {
	var base *Config
	switch preset {
	case PresetCareful:
		base = CarefulConfig()
	default:
		base = DefaultConfig()
	}

	if custom == nil {
		return base
	}

	// Apply overrides if non-zero
	if custom.MouseTimeBase > 0 {
		base.MouseTimeBase = custom.MouseTimeBase
	}
	if custom.MouseTimePerBit > 0 {
		base.MouseTimePerBit = custom.MouseTimePerBit
	}
	if custom.MouseTimeSpread > 0 {
		base.MouseTimeSpread = custom.MouseTimeSpread
	}
	if custom.MouseCurvature > 0 {
		base.MouseCurvature = custom.MouseCurvature
	}
	if custom.MouseOvershootProb > 0 {
		base.MouseOvershootProb = custom.MouseOvershootProb
	}
	if custom.MouseJitterProb > 0 {
		base.MouseJitterProb = custom.MouseJitterProb
	}
	if custom.KeyDelayBase > 0 {
		base.KeyDelayBase = custom.KeyDelayBase
	}
	if custom.KeyDelaySpread > 0 {
		base.KeyDelaySpread = custom.KeyDelaySpread
	}
	if custom.ScrollStepBase > 0 {
		base.ScrollStepBase = custom.ScrollStepBase
	}
	if custom.ScrollDelay > 0 {
		base.ScrollDelay = custom.ScrollDelay
	}
	return base
}
