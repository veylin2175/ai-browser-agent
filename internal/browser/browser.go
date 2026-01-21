package browser

import (
	"fmt"

	"ai-browser-agent/internal/config"
	"github.com/playwright-community/playwright-go"
)

type Browser struct {
	PW      *playwright.Playwright
	Context playwright.BrowserContext
	Page    playwright.Page
}

func Launch(cfg *config.Config) (*Browser, error) {
	if err := playwright.Install(); err != nil {
		return nil, fmt.Errorf("install playwright: %w", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("run playwright: %w", err)
	}

	context, err := pw.Chromium.LaunchPersistentContext(
		cfg.Env.BrowserUserDataDir,
		playwright.BrowserTypeLaunchPersistentContextOptions{
			Headless: playwright.Bool(cfg.Env.BrowserHeadless),
			SlowMo:   playwright.Float(float64(cfg.Env.BrowserSlowMoMs)),
			Viewport: &playwright.Size{
				Width:  cfg.Browser.Viewport.Width,
				Height: cfg.Browser.Viewport.Height,
			},
			Timeout: playwright.Float(float64(cfg.Browser.TimeoutMs)),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("launch context: %w", err)
	}

	var page playwright.Page
	pages := context.Pages()
	if len(pages) > 0 {
		page = pages[0]
	} else {
		page, err = context.NewPage()
		if err != nil {
			return nil, fmt.Errorf("new page: %w", err)
		}
	}

	return &Browser{
		PW:      pw,
		Context: context,
		Page:    page,
	}, nil
}

func (b *Browser) Close() {
	if b.Context != nil {
		_ = b.Context.Close()
	}
	if b.PW != nil {
		_ = b.PW.Stop()
	}
}
