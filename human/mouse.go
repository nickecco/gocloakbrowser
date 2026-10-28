package human

import (
	"context"
	"math"
	"math/rand/v2"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

const (
	frameDuration = 1000.0 / 60.0 // 60fps display frame in ms (~16.6ms)
)

// Point represents a 2D coordinate on the viewport.
type Point struct {
	X float64
	Y float64
}

// Step represents a planned movement waypoint with delay since previous point.
type Step struct {
	X  float64
	Y  float64
	Dt time.Duration
}

func easeProgress(t float64, peak float64) float64 {
	k := math.Log(0.5) / math.Log(peak)
	u := math.Pow(t, k)
	return u * u * u * (10 - 15*u + 6*u*u)
}

func bezier(p0, p1, p2, p3 Point, t float64) Point {
	u := 1 - t
	uu := u * u
	uuu := uu * u
	tt := t * t
	ttt := tt * t

	return Point{
		X: uuu*p0.X + 3*uu*t*p1.X + 3*u*tt*p2.X + ttt*p3.X,
		Y: uuu*p0.Y + 3*uu*t*p1.Y + 3*u*tt*p2.Y + ttt*p3.Y,
	}
}

func randomControlPoints(start, end Point, curvature float64) (Point, Point) {
	dx := end.X - start.X
	dy := end.Y - start.Y
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		dist = 1
	}
	px := -dy / dist
	py := dx / dist

	bias1 := (rand.Float64()*2*curvature - curvature) * dist
	bias2 := bias1 * (0.2 + rand.Float64())
	if rand.Float64() >= 0.8 {
		bias2 = (rand.Float64()*2*curvature - curvature) * dist
	}

	return Point{
		X: start.X + dx*0.25 + px*bias1,
		Y: start.Y + dy*0.25 + py*bias1,
	}, Point{
		X: start.X + dx*0.75 + px*bias2,
		Y: start.Y + dy*0.75 + py*bias2,
	}
}

// PlanStroke plans a single smooth movement between (sx, sy) and (ex, ey).
func PlanStroke(start, end Point, cfg *Config, totalDuration time.Duration) []Step {
	dist := math.Hypot(end.X-start.X, end.Y-start.Y)
	if dist < 0.5 {
		return nil
	}

	totalMs := float64(totalDuration.Milliseconds())
	if totalMs <= 0 {
		// Fitts' Law with log-normal random variation
		medianMs := float64(cfg.MouseTimeBase.Milliseconds()) +
			float64(cfg.MouseTimePerBit.Milliseconds())*math.Log2(1+dist/8.0)
		spread := cfg.MouseTimeSpread
		totalMs = max(frameDuration*2, medianMs*math.Exp(rand.NormFloat64()*spread))
	}

	stepsCount := int(max(5, min(120, math.Round(totalMs/frameDuration))))
	cp1, cp2 := randomControlPoints(start, end, cfg.MouseCurvature)
	peak := 0.22 + rand.Float64()*(0.38-0.22)

	ux, uy := (end.X-start.X)/dist, (end.Y-start.Y)/dist
	amp := min(15.0, dist*0.02) * (rand.Float64()*2 - 1)
	waves := 0.5 + rand.Float64()
	phase := rand.Float64() * math.Pi

	var steps []Step
	lx, ly := start.X, start.Y
	tPrev := 0.0

	for step := range stepsCount {
		i := step + 1
		t := float64(i) / float64(stepsCount)
		if i < stepsCount {
			t = min(1.0, max(0.0, t+(rand.Float64()*0.04-0.02)))
		}

		var x, y float64
		if i == stepsCount {
			x, y = end.X, end.Y
		} else {
			p := bezier(start, cp1, cp2, end, easeProgress(t, peak))
			w := amp * math.Sin(math.Pi*t) * math.Sin(phase+2*math.Pi*waves*t)
			x = math.Round(p.X - uy*w)
			y = math.Round(p.Y + ux*w)
		}

		if math.Abs(x-lx) < 0.5 && math.Abs(y-ly) < 0.5 && i < stepsCount {
			continue
		}

		dtMs := (t - tPrev) * totalMs
		tPrev = t
		lx, ly = x, y

		steps = append(steps, Step{
			X:  x,
			Y:  y,
			Dt: time.Duration(max(1, dtMs)) * time.Millisecond,
		})
	}
	return steps
}

// PlanMove plans a movement from start to end, optionally including an overshoot and correction.
func PlanMove(start, end Point, cfg *Config) []Step {
	dist := math.Hypot(end.X-start.X, end.Y-start.Y)
	if dist < 0.5 {
		return nil
	}

	// Overshoot check for long distances
	if dist > 60 && rand.Float64() < cfg.MouseOvershootProb {
		ang := math.Atan2(end.Y-start.Y, end.X-start.X) + (rand.Float64()*0.2 - 0.1)
		missDist := dist*0.03 + (rand.Float64()*10 + 2)
		overshootPt := Point{
			X: end.X + math.Cos(ang)*missDist,
			Y: end.Y + math.Sin(ang)*missDist,
		}

		firstStroke := PlanStroke(start, overshootPt, cfg, 0)
		pause := Step{
			X:  overshootPt.X,
			Y:  overshootPt.Y,
			Dt: time.Duration(20+rand.IntN(40)) * time.Millisecond,
		}
		secondStroke := PlanStroke(overshootPt, end, cfg, time.Duration(60+rand.IntN(40))*time.Millisecond)

		res := append(firstStroke, pause)
		return append(res, secondStroke...)
	}

	return PlanStroke(start, end, cfg, 0)
}

// MoveMouse performs a humanized mouse move from current position to target coordinate.
func MoveMouse(ctx context.Context, mouse *rod.Mouse, target Point, cfg *Config) error {
	cur := mouse.Position()
	start := Point{X: cur.X, Y: cur.Y}
	steps := PlanMove(start, target, cfg)

	for _, s := range steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if s.Dt > 0 {
			time.Sleep(s.Dt)
		}
		if err := mouse.MoveTo(proto.Point{X: s.X, Y: s.Y}); err != nil {
			return err
		}
	}
	return nil
}
