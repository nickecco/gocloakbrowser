package gocloakbrowser

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/nickecco/gocloakbrowser/human"
)

// Browser wraps rod.Browser with CloakBrowser life-cycle and humanize capabilities.
type Browser struct {
	*rod.Browser

	launcher    *launcher.Launcher
	humanize    bool
	humanConfig *human.Config
}

// NewPage creates a new tab with optional URL and returns a wrapped Page.
func (b *Browser) NewPage(targetURL string) (*Page, error) {
	rodPage, err := b.Browser.Page(proto.TargetCreateTarget{
		URL: targetURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create page: %w", err)
	}

	return &Page{
		Page:        rodPage,
		humanize:    b.humanize,
		humanConfig: b.humanConfig,
	}, nil
}

// MustNewPage creates a page and panics on failure.
func (b *Browser) MustNewPage(targetURL string) *Page {
	p, err := b.NewPage(targetURL)
	if err != nil {
		panic(err)
	}
	return p
}

// WrapPage wraps an existing *rod.Page with CloakBrowser features.
func (b *Browser) WrapPage(p *rod.Page) *Page {
	return &Page{
		Page:        p,
		humanize:    b.humanize,
		humanConfig: b.humanConfig,
	}
}

// Close closes the browser connection and terminates the browser process.
func (b *Browser) Close() error {
	_ = b.Browser.Close()
	if b.launcher != nil {
		b.launcher.Kill()
	}
	return nil
}

// Page wraps rod.Page with human-like interactions.
type Page struct {
	*rod.Page

	humanize    bool
	humanConfig *human.Config
}

// SetHumanize toggles humanized interactions on this page.
func (p *Page) SetHumanize(enable bool) *Page {
	p.humanize = enable
	return p
}

// SetHumanConfig updates the humanize parameters for this page.
func (p *Page) SetHumanConfig(cfg *human.Config) *Page {
	p.humanConfig = cfg
	return p
}

// HumanMove moves the mouse smoothly to the center of the element specified by selector.
func (p *Page) HumanMove(selector string) error {
	el, err := p.Element(selector)
	if err != nil {
		return fmt.Errorf("find element %q: %w", selector, err)
	}
	if err := el.WaitVisible(); err != nil {
		return fmt.Errorf("wait visible %q: %w", selector, err)
	}

	shape, err := el.Shape()
	if err != nil {
		return fmt.Errorf("element shape %q: %w", selector, err)
	}

	pt := shape.OnePointInside()
	if pt == nil {
		box := shape.Box()
		pt = &proto.Point{
			X: box.X + box.Width/2,
			Y: box.Y + box.Height/2,
		}
	}

	return human.MoveMouse(p.GetContext(), p.Mouse, human.Point{X: pt.X, Y: pt.Y}, p.humanConfig)
}

// HumanMoveTo moves the mouse smoothly to arbitrary (x, y) coordinates.
func (p *Page) HumanMoveTo(x, y float64) error {
	return human.MoveMouse(p.GetContext(), p.Mouse, human.Point{X: x, Y: y}, p.humanConfig)
}

// HumanClick moves the mouse with human Bézier curves and clicks the element.
// If humanize is disabled, falls back to direct Click.
func (p *Page) HumanClick(selector string) error {
	el, err := p.Element(selector)
	if err != nil {
		return fmt.Errorf("find element %q: %w", selector, err)
	}
	if err := el.WaitVisible(); err != nil {
		return fmt.Errorf("wait visible %q: %w", selector, err)
	}

	if !p.humanize {
		return el.Click(proto.InputMouseButtonLeft, 1)
	}

	shape, err := el.Shape()
	if err != nil {
		return fmt.Errorf("element shape %q: %w", selector, err)
	}

	pt := shape.OnePointInside()
	if pt == nil {
		box := shape.Box()
		pt = &proto.Point{
			X: box.X + box.Width/2,
			Y: box.Y + box.Height/2,
		}
	}

	// Move along human stroke
	if err := human.MoveMouse(p.GetContext(), p.Mouse, human.Point{X: pt.X, Y: pt.Y}, p.humanConfig); err != nil {
		return err
	}

	// Pre-click hesitation (dwell time: ~30ms - 80ms)
	time.Sleep(time.Duration(30+rand.IntN(50)) * time.Millisecond)

	return p.Mouse.Click(proto.InputMouseButtonLeft, 1)
}

// MustHumanClick executes HumanClick and panics on error.
func (p *Page) MustHumanClick(selector string) *Page {
	if err := p.HumanClick(selector); err != nil {
		panic(err)
	}
	return p
}

// HumanInput moves to element, clicks it to focus, and types text with natural keystroke timing.
func (p *Page) HumanInput(selector, text string) error {
	el, err := p.Element(selector)
	if err != nil {
		return fmt.Errorf("find element %q: %w", selector, err)
	}

	if !p.humanize {
		return el.Input(text)
	}

	// Click to focus
	if err := p.HumanClick(selector); err != nil {
		return err
	}

	// Brief pause after focusing input
	time.Sleep(time.Duration(60+rand.IntN(80)) * time.Millisecond)

	return human.TypeText(p.GetContext(), p.Keyboard, text, p.humanConfig)
}

// MustHumanInput executes HumanInput and panics on error.
func (p *Page) MustHumanInput(selector, text string) *Page {
	if err := p.HumanInput(selector, text); err != nil {
		panic(err)
	}
	return p
}

// HumanScroll performs a smooth, human-like scroll using the mouse wheel.
func (p *Page) HumanScroll(deltaX, deltaY float64) error {
	if !p.humanize {
		return p.Mouse.Scroll(deltaX, deltaY, 5)
	}
	return human.Scroll(p.GetContext(), p.Mouse, deltaX, deltaY, p.humanConfig)
}

// MustHumanScroll executes HumanScroll and panics on error.
func (p *Page) MustHumanScroll(deltaX, deltaY float64) *Page {
	if err := p.HumanScroll(deltaX, deltaY); err != nil {
		panic(err)
	}
	return p
}
