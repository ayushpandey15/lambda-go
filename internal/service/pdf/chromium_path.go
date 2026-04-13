package pdf

import (
	"os"
	"strings"
)

// ChromiumPath resolves the Chrome/Chromium binary for chromedp and rod.
//
// Order:
//  1. CHROME_PATH — set this in Lambda when your layer uses a non-standard layout.
//  2. CHROMIUM_PATH — same intent, alternate name used by some tooling.
//  3. Common Lambda layer paths under /opt, then typical Linux locations.
//
// For rod only, ROD_BROWSER_PATH is checked first (see rodChromiumBin in rod.go).
func ChromiumPath() string {
	for _, k := range []string{"CHROME_PATH", "CHROMIUM_PATH"} {
		if p := strings.TrimSpace(os.Getenv(k)); p != "" {
			return p
		}
	}
	for _, p := range chromiumPathCandidates() {
		if isExecutableFile(p) {
			return p
		}
	}
	return ""
}

func chromiumPathCandidates() []string {
	return []string{
		"/opt/chromium/chromium",
		"/opt/chrome/chrome",
		"/opt/google/chrome/chrome",
		"/opt/google/chrome/google-chrome",
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
	}
}

func isExecutableFile(path string) bool {
	fi, err := os.Stat(path)
	if err != nil || !fi.Mode().IsRegular() {
		return false
	}
	return fi.Mode().Perm()&0111 != 0
}
