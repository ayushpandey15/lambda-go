package pdf

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ConvertChromedp renders HTML or loads URL and returns PDF bytes using headless Chrome (chromedp).
func ConvertChromedp(ctx context.Context, in Input) ([]byte, error) {
	if in.HTML == "" && in.URL == "" {
		return nil, fmt.Errorf("chromedp: html or url required")
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.NoSandbox,
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	navURL := in.URL
	if navURL == "" {
		navURL = "data:text/html;base64," + base64.StdEncoding.EncodeToString([]byte(in.HTML))
	}

	var pdf []byte
	tasks := chromedp.Tasks{
		chromedp.Navigate(navURL),
		chromedp.Sleep(300 * time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				Do(ctx)
			return err
		}),
	}

	if err := chromedp.Run(taskCtx, tasks); err != nil {
		return nil, fmt.Errorf("chromedp: %w", err)
	}
	return pdf, nil
}
