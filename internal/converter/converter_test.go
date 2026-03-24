package converter_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hiro/pdf2png/internal/converter"
)

// minimalPDF is a minimal valid 1-page PDF that MuPDF can parse.
// Byte offsets in the xref table are exact for this specific byte sequence.
var minimalPDF = []byte(
	"%PDF-1.0\n" +
		"1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n" +
		"2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj\n" +
		"3 0 obj<</Type/Page/MediaBox[0 0 72 72]>>endobj\n" +
		"xref\n" +
		"0 4\n" +
		"0000000000 65535 f \n" +
		"0000000009 00000 n \n" +
		"0000000052 00000 n \n" +
		"0000000101 00000 n \n" +
		"trailer<</Size 4/Root 1 0 R>>\n" +
		"startxref\n" +
		"149\n" +
		"%%EOF")

func writeTempPDF(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "test_*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.Write(minimalPDF); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func TestConvert_SinglePage(t *testing.T) {
	pdfPath := writeTempPDF(t)
	outDir := t.TempDir()

	files, err := converter.Convert(pdfPath, outDir)
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// 1-page PDF → 1 file
	if len(files) != 1 {
		t.Errorf("got %d files, want 1", len(files))
	}

	// File must exist and be non-empty
	for _, p := range files {
		info, err := os.Stat(p)
		if err != nil {
			t.Errorf("stat %q: %v", p, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("file %q is empty", p)
		}
	}
}

func TestConvert_FileNaming(t *testing.T) {
	pdfPath := writeTempPDF(t)
	outDir := t.TempDir()

	files, err := converter.Convert(pdfPath, outDir)
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	// File name must be page_*.png
	if len(files) < 1 {
		t.Fatal("no files returned")
	}
	base := filepath.Base(files[0])
	if !strings.HasPrefix(base, "page_") || !strings.HasSuffix(base, ".png") {
		t.Errorf("unexpected filename: %q", base)
	}
}

func TestConvert_InvalidPath(t *testing.T) {
	_, err := converter.Convert("/nonexistent/path/file.pdf", t.TempDir())
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}
