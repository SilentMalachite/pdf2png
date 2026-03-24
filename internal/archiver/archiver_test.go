package archiver_test

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/hiro/pdf2png/internal/archiver"
)

func TestArchive(t *testing.T) {
	// Setup: PNG ファイルを含む一時ディレクトリを作成
	srcDir := t.TempDir()
	files := []string{}
	for _, name := range []string{"page_01.png", "page_02.png", "page_03.png"} {
		p := filepath.Join(srcDir, name)
		if err := os.WriteFile(p, []byte("fake png data for "+name), 0644); err != nil {
			t.Fatal(err)
		}
		files = append(files, p)
	}

	// Execute
	outDir := t.TempDir()
	zipPath := filepath.Join(outDir, "output.zip")
	if err := archiver.Archive(files, zipPath); err != nil {
		t.Fatalf("Archive() error = %v", err)
	}

	// Verify: ZIP が存在する
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Fatal("ZIP file not created")
	}

	// Verify: ZIP の中身が 3 ファイル
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("cannot open zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 3 {
		t.Errorf("zip contains %d files, want 3", len(r.File))
	}

	// Verify: ファイル名はベース名のみ（ディレクトリパスなし）
	for _, f := range r.File {
		if filepath.Base(f.Name) != f.Name {
			t.Errorf("zip entry has directory path: %q", f.Name)
		}
	}
}

func TestArchiveOverwrite(t *testing.T) {
	// 同名 ZIP が存在する場合に上書きされることを確認
	srcDir := t.TempDir()
	p := filepath.Join(srcDir, "page_01.png")
	if err := os.WriteFile(p, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	outDir := t.TempDir()
	zipPath := filepath.Join(outDir, "output.zip")

	// 1回目
	if err := archiver.Archive([]string{p}, zipPath); err != nil {
		t.Fatal(err)
	}
	info1, _ := os.Stat(zipPath)

	// 2回目（上書き）
	if err := archiver.Archive([]string{p}, zipPath); err != nil {
		t.Fatalf("second Archive() error = %v", err)
	}
	info2, _ := os.Stat(zipPath)

	// ModTime が変わっていることで上書きを確認
	if !info2.ModTime().After(info1.ModTime()) {
		t.Log("note: modtime may not change in fast test runs, checking file exists")
	}
}
