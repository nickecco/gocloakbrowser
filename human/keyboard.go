package human

import (
	"context"
	"math"
	"math/rand/v2"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

// TypeText types text character-by-character with natural, human-like timings.
func TypeText(ctx context.Context, kb *rod.Keyboard, text string, cfg *Config) error {
	baseMs := float64(cfg.KeyDelayBase.Milliseconds())
	spread := cfg.KeyDelaySpread

	for _, r := range text {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Press keydown
		key := input.Key(r)
		if err := kb.Press(key); err != nil {
			return err
		}

		// Dwell time (how long the physical key remains depressed: ~25ms-70ms)
		dwellMs := 25.0 + rand.Float64()*45.0
		time.Sleep(time.Duration(dwellMs) * time.Millisecond)

		// Release keyup
		if err := kb.Release(key); err != nil {
			return err
		}

		// Flight time (pause until the next key is struck)
		flightMs := math.Max(15.0, baseMs*(1.0+rand.NormFloat64()*spread))

		// Natural pause after punctuation or space
		switch r {
		case ' ', '\t':
			flightMs += 30.0 + rand.Float64()*60.0
		case '.', ',', '!', '?', ';', ':':
			flightMs += 80.0 + rand.Float64()*120.0
		case '\n':
			flightMs += 120.0 + rand.Float64()*250.0
		}

		time.Sleep(time.Duration(flightMs) * time.Millisecond)
	}
	return nil
}
