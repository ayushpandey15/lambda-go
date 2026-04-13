package pdf

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/ayushpandey15/lambda-go/internal/pkg/requestbody"
	"github.com/ayushpandey15/lambda-go/internal/pkg/response"
	pdfsvc "github.com/ayushpandey15/lambda-go/internal/service/pdf"

	"github.com/gin-gonic/gin"
)

type convertRequest struct {
	ConvertType string `json:"convertType" validate:"required,oneof=chromedp rod wkhtml gotenberg"`
	HTMLData    string `json:"htmlData" validate:"required"`
}

// HTMLToPDF handles POST /html-to-pdf.
// Body: convertType "chromedp" | "rod" | "wkhtml" | "gotenberg", htmlData = base64-encoded HTML.
// chromedp and rod use headless Chromium in-process (Lambda: custom image/layer). gotenberg uses GOTENBERG_URL (remote service).
// Response: JSON success with data.pdfData = standard base64-encoded PDF.
func HTMLToPDF(c *gin.Context) {
	req, ok := requestbody.BindJSON(c, func(r *convertRequest) {
		r.ConvertType = strings.ToLower(strings.TrimSpace(r.ConvertType))
	})
	if !ok {
		return
	}

	ct := req.ConvertType

	htmlPlain, decErr := decodeBase64HTML(req.HTMLData)
	if decErr != nil {
		response.WriteErrorResponse(c, response.ErrInvalidInput.WithMessage("htmlData must be valid base64"))
		return
	}
	htmlPlain = strings.TrimSpace(htmlPlain)
	if htmlPlain == "" {
		response.WriteErrorResponse(c, response.ErrInvalidInput.WithMessage("htmlData decodes to empty HTML"))
		return
	}

	in := pdfsvc.Input{HTML: htmlPlain}

	var (
		out []byte
		err error
	)
	switch ct {
	case "chromedp":
		out, err = pdfsvc.ConvertChromedp(c.Request.Context(), in)
	case "rod":
		out, err = pdfsvc.ConvertRod(c.Request.Context(), in)
	// case "wkhtml":
	// 	out, err = pdfsvc.ConvertWkhtmltopdf(c.Request.Context(), in)
	// case "gotenberg":
	// 	out, err = pdfsvc.ConvertGotenberg(c.Request.Context(), in)
	// }
	default:
		response.WriteErrorResponse(c, response.ErrInvalidInput.WithMessage("convertType must be one of: chromedp, rod, wkhtml, gotenberg"))
		return
	}

	if err != nil {
		response.WriteErrorResponse(c, response.ErrInternal.WithMessage(err.Error()))
		return
	}

	pdfB64 := base64.StdEncoding.EncodeToString(out)
	response.WriteSucessResponse(c, "pdf generated", gin.H{"pdfData": pdfB64})
}

// decodeBase64HTML decodes base64 payload (standard, raw, or URL-safe variants).
func decodeBase64HTML(encoded string) (string, error) {
	s := strings.TrimSpace(encoded)
	if s == "" {
		return "", errors.New("empty")
	}
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")

	for _, enc := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		raw, err := enc.DecodeString(s)
		if err == nil {
			return string(raw), nil
		}
	}
	return "", errors.New("invalid base64")
}
