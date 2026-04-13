package pdf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gowk "github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

// resolveWkhtmltopdfBin finds the wkhtmltopdf executable for AWS Lambda and local machines.
//
// Lambda always runs Linux (e.g. Amazon Linux 2 / 2023). The managed runtime does not ship
// wkhtmltopdf; you normally bundle it under LAMBDA_TASK_ROOT or set WKHTMLTOPDF_PATH.
// Extra absolute paths below are tried only if the file exists—on Lambda, macOS Homebrew
// paths are harmless no-ops; on Linux, /opt/wkhtmltox and /usr/bin often come from layers, Docker, or apt/yum.
//
// Order: WKHTMLTOPDF_PATH → PATH → LAMBDA_TASK_ROOT relative → common install locations.
func resolveWkhtmltopdfBin() (string, error) {
	if p := strings.TrimSpace(os.Getenv("WKHTMLTOPDF_PATH")); p != "" {
		if err := statRegularFile(p); err != nil {
			return "", fmt.Errorf("WKHTMLTOPDF_PATH %q: %w", p, err)
		}
		return p, nil
	}
	if p, err := exec.LookPath("wkhtmltopdf"); err == nil {
		return p, nil
	}
	if root := strings.TrimSpace(os.Getenv("LAMBDA_TASK_ROOT")); root != "" {
		for _, rel := range []string{"wkhtmltopdf", filepath.Join("bin", "wkhtmltopdf")} {
			p := filepath.Join(root, rel)
			if err := statRegularFile(p); err == nil {
				return p, nil
			}
		}
	}
	for _, p := range []string{
		"/opt/wkhtmltox/bin/wkhtmltopdf", // official Linux .tar.xz layout; Docker / many Lambda images
		"/usr/local/bin/wkhtmltopdf",     // manual install; Intel Homebrew (macOS)
		"/usr/bin/wkhtmltopdf",           // apt/yum on Linux
		"/opt/homebrew/bin/wkhtmltopdf",  // Apple Silicon Homebrew (macOS local dev)
	} {
		if err := statRegularFile(p); err == nil {
			return p, nil
		}
	}
	return "", errors.New(`wkhtmltopdf not found: set WKHTMLTOPDF_PATH (recommended on Lambda), bundle under LAMBDA_TASK_ROOT, install on PATH or under /opt/wkhtmltox, or use convertType "chromedp"`)
}

func statRegularFile(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return errors.New("not a file")
	}
	return nil
}

// ConvertWkhtmltopdf renders HTML or a URL to PDF using wkhtmltopdf via
// github.com/SebastiaanKlippert/go-wkhtmltopdf (still requires the wkhtmltopdf binary).
func ConvertWkhtmltopdf(ctx context.Context, in Input) ([]byte, error) {
	if in.HTML == "" && in.URL == "" {
		return nil, fmt.Errorf("wkhtmltopdf: html or url required")
	}

	bin, err := resolveWkhtmltopdfBin()
	if err != nil {
		return nil, err
	}
	gowk.SetPath(bin)

	pdfg, err := gowk.NewPDFGenerator()
	if err != nil {
		return nil, err
	}

	pdfg.Quiet.Set(true)

	if in.URL != "" {
		page := gowk.NewPage(in.URL)
		page.EnableLocalFileAccess.Set(true)
		pdfg.AddPage(page)
	} else {
		page := gowk.NewPageReader(bytes.NewReader([]byte(in.HTML)))
		page.EnableLocalFileAccess.Set(true)
		pdfg.AddPage(page)
	}

	if err := pdfg.CreateContext(ctx); err != nil {
		return nil, fmt.Errorf("wkhtmltopdf: %w", err)
	}

	out := pdfg.Bytes()
	if len(out) == 0 {
		return nil, fmt.Errorf("wkhtmltopdf: empty pdf output")
	}
	return out, nil
}
