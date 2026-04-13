package pdf

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	pdfLimitOnce sync.Once
	pdfSem       chan struct{} // buffered; acquire sends, release receives
)

// pdfMaxConcurrent returns PDF_MAX_CONCURRENT from the environment.
// 0 or unset means no limit. Values < 0 are treated as 0.
func pdfMaxConcurrent() int {
	s := strings.TrimSpace(os.Getenv("PDF_MAX_CONCURRENT"))
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func pdfSemInit() {
	pdfLimitOnce.Do(func() {
		n := pdfMaxConcurrent()
		if n > 0 {
			pdfSem = make(chan struct{}, n)
		}
	})
}

// pdfAcquire blocks until a PDF slot is available or ctx is done.
func pdfAcquire(ctx context.Context) error {
	pdfSemInit()
	if pdfSem == nil {
		return nil
	}
	select {
	case pdfSem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func pdfRelease() {
	pdfSemInit()
	if pdfSem == nil {
		return
	}
	<-pdfSem
}
