package human_test

import (
	"math"
	"testing"
	"time"

	"github.com/nickecco/gocloakbrowser/human"
)

func TestPlanStroke(t *testing.T) {
	cfg := human.DefaultConfig()
	start := human.Point{X: 100, Y: 100}
	end := human.Point{X: 500, Y: 400}

	steps := human.PlanStroke(start, end, cfg, 200*time.Millisecond)
	if len(steps) < 5 {
		t.Fatalf("expected at least 5 steps in stroke, got %d", len(steps))
	}

	last := steps[len(steps)-1]
	if math.Abs(last.X-end.X) > 1 || math.Abs(last.Y-end.Y) > 1 {
		t.Errorf("last step (%f, %f) did not reach destination (%f, %f)", last.X, last.Y, end.X, end.Y)
	}

	// Verify monotonic or non-negative durations
	for i, s := range steps {
		if s.Dt < 0 {
			t.Errorf("step %d has negative duration: %v", i, s.Dt)
		}
	}
}

func TestPlanMoveOvershoot(t *testing.T) {
	cfg := human.DefaultConfig()
	cfg.MouseOvershootProb = 1.0 // force overshoot for long distance

	start := human.Point{X: 50, Y: 50}
	end := human.Point{X: 600, Y: 600}

	steps := human.PlanMove(start, end, cfg)
	if len(steps) == 0 {
		t.Fatalf("expected movement steps")
	}

	last := steps[len(steps)-1]
	if math.Abs(last.X-end.X) > 1 || math.Abs(last.Y-end.Y) > 1 {
		t.Errorf("last step (%f, %f) did not reach destination (%f, %f)", last.X, last.Y, end.X, end.Y)
	}
}
