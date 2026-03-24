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

// zeroPDF is a minimal PDF with 0 pages.
var zeroPDF = []byte(
	"%PDF-1.0\n" +
		"1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n" +
		"2 0 obj<</Type/Pages/Kids[]/Count 0>>endobj\n" +
		"xref\n" +
		"0 3\n" +
		"0000000000 65535 f \n" +
		"0000000009 00000 n \n" +
		"0000000052 00000 n \n" +
		"trailer<</Size 3/Root 1 0 R>>\n" +
		"startxref\n" +
		"98\n" +
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
	// 1-page PDF: padding width = len("1") = 1, so filename is page_1.png (no zero padding)
	const want = "page_1.png"
	if base != want {
		t.Errorf("filename = %q, want %q", base, want)
	}
}

func TestConvert_NoPages(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "zero_*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(zeroPDF); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	_, err = converter.Convert(f.Name(), t.TempDir())
	if err == nil {
		t.Fatal("expected error for 0-page PDF, got nil")
	}
	if !strings.Contains(err.Error(), "no pages") {
		t.Errorf("error = %q, want to contain 'no pages'", err.Error())
	}
}

func TestConvert_PasswordProtected(t *testing.T) {
	// This test requires a password-protected PDF fixture.
	// To create one on macOS:
	//   1. Open any PDF in Preview
	//   2. File → Export as PDF → Security Options → set a password
	//   3. Save as testdata/encrypted.pdf
	const fixture = "testdata/encrypted.pdf"
	if _, err := os.Stat(fixture); os.IsNotExist(err) {
		t.Skip("testdata/encrypted.pdf not found — see comment for how to create it")
	}

	_, err := converter.Convert(fixture, t.TempDir())
	if err == nil {
		t.Fatal("expected error for password-protected PDF, got nil")
	}
	if !strings.Contains(err.Error(), "password") {
		t.Errorf("error = %q, want to contain 'password'", err.Error())
	}
}

func TestConvert_InvalidPath(t *testing.T) {
	_, err := converter.Convert("/nonexistent/path/file.pdf", t.TempDir())
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}
