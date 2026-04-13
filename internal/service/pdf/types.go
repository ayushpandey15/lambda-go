// Package pdf converts HTML to PDF using chromedp, rod, wkhtmltopdf, or a remote Gotenberg service,
// selected by the API JSON field convertType.
package pdf

// Input is shared input for HTML→PDF converters (HTML is decoded plain text, not base64).
type Input struct {
	HTML string
	URL  string
	// BaseURL is optional. For wkhtmltopdf with HTML from stdin, set this to the page origin
	// (e.g. https://example.com/) so relative <img src="/a.png"> and CSS urls resolve.
	BaseURL string
}
