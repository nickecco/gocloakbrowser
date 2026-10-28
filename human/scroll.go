package human

import (
	"context"
	"math"
	"math/rand/v2"
	"time"

	"github.com/go-rod/rod"
)

// Scroll performs a smooth, burst-like scroll simulating a human mouse wheel.
func Scroll(ctx context.Context, mouse *rod.Mouse, deltaX, deltaY float64, cfg *Config) error {
	totalDist := math.Hypot(deltaX, deltaY)
	if totalDist == 0 {
		return nil
	}

	stepBase := float64(cfg.ScrollStepBase)
	bursts := int(max(1, math.Round(totalDist/stepBase)))

	stepX := deltaX / float64(bursts)
	stepY := deltaY / float64(bursts)

	for range bursts {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Micro jitter in wheel notch magnitude
		jitter := 0.85 + rand.Float64()*0.3
		curX := stepX * jitter
		curY := stepY * jitter

		// Dispatch 2-4 sub-steps per wheel notch for browser smoothness
		subSteps := 2 + rand.IntN(3)
		if err := mouse.Scroll(curX, curY, subSteps); err != nil {
			return err
		}

		delay := cfg.ScrollDelay + time.Duration(rand.IntN(15))*time.Millisecond
		time.Sleep(delay)
	}
	return nil
}
