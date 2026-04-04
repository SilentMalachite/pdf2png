package archiver_test

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/SilentMalachite/pdf2png/internal/archiver"
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
	// 同名 ZIP が存在する場合に内容が置き換わることを確認
	srcDir := t.TempDir()
	outDir := t.TempDir()
	zipPath := filepath.Join(outDir, "output.zip")

	// 1回目: page_first.png を含む ZIP を作成
	first := filepath.Join(srcDir, "page_first.png")
	if err := os.WriteFile(first, []byte("first"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := archiver.Archive([]string{first}, zipPath); err != nil {
		t.Fatal(err)
	}

	// 2回目: page_second.png で上書き
	second := filepath.Join(srcDir, "page_second.png")
	if err := os.WriteFile(second, []byte("second"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := archiver.Archive([]string{second}, zipPath); err != nil {
		t.Fatalf("second Archive() error = %v", err)
	}

	// ZIP の内容が 2 回目の内容に完全に置き換わっていることを確認
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("cannot open zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 1 {
		t.Errorf("zip contains %d files after overwrite, want 1", len(r.File))
	}
	if len(r.File) > 0 && r.File[0].Name != "page_second.png" {
		t.Errorf("zip entry name = %q, want page_second.png", r.File[0].Name)
	}
}
