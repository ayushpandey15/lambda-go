package pdf

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

var (
	rodMu      sync.Mutex
	rodBrowser *rod.Browser
	rodLaunch  *launcher.Launcher
)

func rodLauncherBase(ctx context.Context) *launcher.Launcher {
	l := launcher.New().
		Context(ctx).
		HeadlessNew(true).
		NoSandbox(true).
		Leakless(false).
		Set("disable-gpu", "true").
		Set("disable-extensions", "true").
		Set("mute-audio", "true")
	if p := chromiumExecutableForRod(); p != "" {
		l = l.Bin(p)
	}
	return l
}

// chromiumExecutableForRod prefers ROD_BROWSER_PATH, then CHROME_PATH / CHROMIUM_PATH (Lambda/container).
func chromiumExecutableForRod() string {
	if p := strings.TrimSpace(os.Getenv("ROD_BROWSER_PATH")); p != "" {
		return p
	}
	for _, k := range []string{"CHROME_PATH", "CHROMIUM_PATH"} {
		if p := strings.TrimSpace(os.Getenv(k)); p != "" {
			return p
		}
	}
	return ""
}

func rodReset() {
	if rodBrowser != nil {
		_ = rodBrowser.Close()
		rodBrowser = nil
	}
	if rodLaunch != nil {
		rodLaunch.Cleanup()
		rodLaunch = nil
	}
}

func rodEnsureBrowser() error {
	if rodBrowser != nil {
		return nil
	}
	l := rodLauncherBase(context.Background())
	ws, err := l.Launch()
	if err != nil {
		return err
	}
	b := rod.New().Context(context.Background()).ControlURL(ws)
	if err := b.Connect(); err != nil {
		l.Cleanup()
		return err
	}
	rodLaunch = l
	rodBrowser = b
	return nil
}

func rodAfterLoadPause() {
	ms := strings.TrimSpace(os.Getenv("ROD_AFTER_LOAD_SLEEP_MS"))
	if ms == "" {
		return
	}
	n, err := strconv.Atoi(ms)
	if err != nil || n <= 0 {
		return
	}
	time.Sleep(time.Duration(n) * time.Millisecond)
}

func convertRodCore(ctx context.Context, browser *rod.Browser, in Input) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	navURL := in.URL
	if navURL == "" {
		navURL = "data:text/html;base64," + base64.StdEncoding.EncodeToString([]byte(in.HTML))
	}

	page, err := browser.Page(proto.TargetCreateTarget{URL: navURL})
	if err != nil {
		return nil, fmt.Errorf("rod: page: %w", err)
	}
	defer func() { _ = page.Close() }()

	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("rod: wait load: %w", err)
	}
	rodAfterLoadPause()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	stream, err := page.PDF(&proto.PagePrintToPDF{PrintBackground: true})
	if err != nil {
		return nil, fmt.Errorf("rod: pdf: %w", err)
	}
	defer stream.Close()

	pdf, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("rod: read pdf stream: %w", err)
	}
	if len(pdf) == 0 {
		return nil, fmt.Errorf("rod: empty pdf")
	}
	return pdf, nil
}

// ConvertRod renders HTML or loads a URL to PDF using headless Chromium via go-rod/rod.
//
// By default it reuses one browser process for the lifetime of the process (much faster on warm
// Lambda / long-running servers). Set ROD_NEW_BROWSER_EACH_REQUEST=1 to launch and tear down
// Chrome on every call (slower, stronger isolation).
//
// Optional ROD_AFTER_LOAD_SLEEP_MS: extra milliseconds after DOM load (e.g. 200) if lazy images need it;
// default is none for speed.
func ConvertRod(ctx context.Context, in Input) ([]byte, error) {
	if in.HTML == "" && in.URL == "" {
		return nil, fmt.Errorf("rod: html or url required")
	}

	if strings.TrimSpace(os.Getenv("ROD_NEW_BROWSER_EACH_REQUEST")) == "1" {
		return convertRodIsolated(ctx, in)
	}

	rodMu.Lock()
	defer rodMu.Unlock()

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if err := rodEnsureBrowser(); err != nil {
			return nil, fmt.Errorf("rod: browser: %w", err)
		}
		b := rodBrowser.Context(ctx)
		pdf, err := convertRodCore(ctx, b, in)
		if err == nil {
			return pdf, nil
		}
		lastErr = err
		rodReset()
	}
	return nil, lastErr
}

func convertRodIsolated(ctx context.Context, in Input) ([]byte, error) {
	l := rodLauncherBase(ctx)
	ws, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("rod: launch: %w", err)
	}
	defer l.Cleanup()

	b := rod.New().Context(ctx).ControlURL(ws)
	if err := b.Connect(); err != nil {
		return nil, fmt.Errorf("rod: connect: %w", err)
	}
	defer func() { _ = b.Close() }()

	return convertRodCore(ctx, b, in)
}
