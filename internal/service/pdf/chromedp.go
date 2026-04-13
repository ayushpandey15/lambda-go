package pdf

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

var (
	chromedpOnce     sync.Once
	chromedpAllocCtx context.Context
)

func chromedpAllocatorOpts() []chromedp.ExecAllocatorOption {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// Override default --headless with Chromium’s newer headless implementation.
		chromedp.Flag("headless", "new"),
		chromedp.NoSandbox,
		chromedp.DisableGPU,
	)
	if p := chromiumExecutable(); p != "" {
		opts = append(opts, chromedp.ExecPath(p))
	}
	return opts
}

func chromedpEnsureAllocator() {
	chromedpOnce.Do(func() {
		opts := chromedpAllocatorOpts()
		var cancel context.CancelFunc
		chromedpAllocCtx, cancel = chromedp.NewExecAllocator(context.Background(), opts...)
		_ = cancel // keep browser alive for reuse across requests; do not call cancel in normal operation
	})
}

// contextMergeCancel returns a child of parent that is cancelled when either parent or whenDone is cancelled.
func contextMergeCancel(parent, whenDone context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	go func() {
		select {
		case <-whenDone.Done():
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}

func chromedpAfterLoadDuration() time.Duration {
	ms := strings.TrimSpace(os.Getenv("CHROMEDP_AFTER_LOAD_SLEEP_MS"))
	if ms == "" {
		return 0
	}
	n, err := strconv.Atoi(ms)
	if err != nil || n <= 0 {
		return 0
	}
	return time.Duration(n) * time.Millisecond
}

// ConvertChromedp renders HTML or loads URL and returns PDF bytes using headless Chrome (chromedp).
//
// By default it reuses one browser process (fast on warm Lambda). Set CHROMEDP_NEW_BROWSER_EACH_REQUEST=1
// to launch and tear down Chrome on every call (slower, stronger isolation).
//
// Optional CHROMEDP_AFTER_LOAD_SLEEP_MS: extra wait after navigation for lazy content (default 0 for speed).
func ConvertChromedp(ctx context.Context, in Input) ([]byte, error) {
	if in.HTML == "" && in.URL == "" {
		return nil, fmt.Errorf("chromedp: html or url required")
	}

	if strings.TrimSpace(os.Getenv("CHROMEDP_NEW_BROWSER_EACH_REQUEST")) == "1" {
		return convertChromedpIsolated(ctx, in)
	}

	chromedpEnsureAllocator()

	parent, cancelParent := contextMergeCancel(chromedpAllocCtx, ctx)
	defer cancelParent()

	taskCtx, cancelTask := chromedp.NewContext(parent)
	defer cancelTask()

	return chromedpRunPDF(taskCtx, in)
}

func convertChromedpIsolated(ctx context.Context, in Input) ([]byte, error) {
	opts := chromedpAllocatorOpts()
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	return chromedpRunPDF(taskCtx, in)
}

func chromedpRunPDF(taskCtx context.Context, in Input) ([]byte, error) {
	var pdf []byte
	var tasks chromedp.Tasks

	if in.URL != "" {
		tasks = chromedp.Tasks{chromedp.Navigate(in.URL)}
	} else {
		// about:blank + setDocumentContent avoids data: URL base64 and large-URL parsing; often faster than Navigate(data:...).
		html := in.HTML
		tasks = chromedp.Tasks{
			chromedp.Navigate("about:blank"),
			chromedp.ActionFunc(func(ctx context.Context) error {
				frameTree, err := page.GetFrameTree().Do(ctx)
				if err != nil {
					return err
				}
				return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
			}),
			chromedp.Poll(`document.readyState === 'complete'`, nil),
		}
	}

	if d := chromedpAfterLoadDuration(); d > 0 {
		tasks = append(tasks, chromedp.Sleep(d))
	}
	tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		pdf, _, err = page.PrintToPDF().
			WithPrintBackground(true).
			Do(ctx)
		return err
	}))

	if err := chromedp.Run(taskCtx, tasks); err != nil {
		return nil, fmt.Errorf("chromedp: %w", err)
	}
	return pdf, nil
}

// chromiumExecutable returns CHROME_PATH or CHROMIUM_PATH when set (typical on Lambda/container images).
func chromiumExecutable() string {
	for _, k := range []string{"CHROME_PATH", "CHROMIUM_PATH"} {
		if p := strings.TrimSpace(os.Getenv(k)); p != "" {
			return p
		}
	}
	return ""
}
