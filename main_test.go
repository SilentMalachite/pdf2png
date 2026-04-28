package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConvertToZipRejectsExistingOutputWithoutOverwrite(t *testing.T) {
	pdfDir := t.TempDir()
	outDir := t.TempDir()

	pdfPath := filepath.Join(pdfDir, "report.pdf")
	if err := os.WriteFile(pdfPath, []byte("placeholder"), 0644); err != nil {
		t.Fatal(err)
	}

	zipPath := filepath.Join(outDir, "report.zip")
	if err := os.WriteFile(zipPath, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	gotPath, err := convertToZip(pdfPath, outDir, nil, false)
	if err == nil {
		t.Fatal("expected existing output error, got nil")
	}
	if !errors.Is(err, ErrOutputExists) {
		t.Fatalf("error = %v, want ErrOutputExists", err)
	}
	if gotPath != zipPath {
		t.Fatalf("zip path = %q, want %q", gotPath, zipPath)
	}
}

func TestConvertToZipValidatesOutputDirectory(t *testing.T) {
	pdfDir := t.TempDir()
	pdfPath := filepath.Join(pdfDir, "report.pdf")
	if err := os.WriteFile(pdfPath, []byte("placeholder"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := convertToZip(pdfPath, filepath.Join(t.TempDir(), "missing"), nil, false)
	if err == nil {
		t.Fatal("expected output directory error, got nil")
	}
	if !strings.Contains(err.Error(), "cannot access output directory") {
		t.Fatalf("error = %q, want output directory message", err.Error())
	}
}
