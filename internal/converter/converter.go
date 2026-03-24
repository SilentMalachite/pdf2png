package converter

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	fitz "github.com/gen2brain/go-fitz"
)

// Convert converts each page of the PDF at pdfPath to a PNG file in outDir.
// Returns the list of created PNG file paths in page order.
// Uses 300 DPI (scale = 300/72 ≈ 4.167).
func Convert(pdfPath, outDir string) ([]string, error) {
	doc, err := fitz.New(pdfPath)
	if err != nil {
		if isPasswordError(err) {
			return nil, fmt.Errorf("PDF is password-protected")
		}
		return nil, fmt.Errorf("failed to open PDF: %w", err)
	}
	defer doc.Close()

	n := doc.NumPage()
	if n == 0 {
		return nil, fmt.Errorf("PDF has no pages")
	}

	// Zero-pad width based on total page count
	digits := len(strconv.Itoa(n))
	format := fmt.Sprintf("page_%%0%dd.png", digits)

	var files []string
	for i := 0; i < n; i++ {
		fmt.Printf("Converting page %d/%d...\n", i+1, n)

		img, err := doc.ImageDPI(i, 300)
		if err != nil {
			return nil, fmt.Errorf("failed to convert page %d: %w", i+1, err)
		}

		name := fmt.Sprintf(format, i+1)
		outPath := filepath.Join(outDir, name)

		f, err := os.Create(outPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s: %w", name, err)
		}

		if err := png.Encode(f, img); err != nil {
			f.Close()
			return nil, fmt.Errorf("failed to encode page %d: %w", i+1, err)
		}
		f.Close()

		files = append(files, outPath)
	}

	return files, nil
}

func isPasswordError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "password") || strings.Contains(msg, "encrypted")
}
