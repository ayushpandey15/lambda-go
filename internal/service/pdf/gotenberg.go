package pdf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

// ConvertGotenberg sends HTML (or a page URL) to a Gotenberg 8+ instance and returns PDF bytes.
// Set GOTENBERG_URL to the service root (e.g. https://gotenberg.internal:3000 or http://localhost:3000).
// No Chromium binary is required inside this process; Gotenberg runs Chrome in its container.
//
// See: https://gotenberg.dev/docs/getting-started/introduction
func ConvertGotenberg(ctx context.Context, in Input) ([]byte, error) {
	base := strings.TrimSuffix(strings.TrimSpace(os.Getenv("GOTENBERG_URL")), "/")
	if base == "" {
		return nil, fmt.Errorf("gotenberg: set GOTENBERG_URL to your Gotenberg base URL (e.g. http://gotenberg:3000)")
	}

	if in.URL != "" {
		return gotenbergChromiumConvertURL(ctx, base, in.URL)
	}
	if in.HTML == "" {
		return nil, fmt.Errorf("gotenberg: html or url required")
	}
	return gotenbergChromiumConvertHTML(ctx, base, in.HTML)
}

func gotenbergChromiumConvertHTML(ctx context.Context, base, html string) ([]byte, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, err
	}
	if _, err := io.WriteString(part, html); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/forms/chromium/convert/html", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	return doGotenberg(req)
}

func gotenbergChromiumConvertURL(ctx context.Context, base, pageURL string) ([]byte, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if err := w.WriteField("url", pageURL); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/forms/chromium/convert/url", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	return doGotenberg(req)
}

func doGotenberg(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gotenberg: request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		msg := string(body)
		if len(msg) > 800 {
			msg = msg[:800] + "…"
		}
		return nil, fmt.Errorf("gotenberg: %s: %s", resp.Status, msg)
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("gotenberg: empty PDF response")
	}
	return body, nil
}
