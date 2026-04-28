package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/SilentMalachite/pdf2png/internal/archiver"
	"github.com/SilentMalachite/pdf2png/internal/converter"
)

func main() {
	if len(os.Args) == 1 {
		runGUI()
		return
	}

	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: pdf2png <file.pdf>")
		pause()
		os.Exit(1)
	}

	if err := run(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		pause()
		os.Exit(1)
	}

	pause()
}

var ErrOutputExists = errors.New("output ZIP already exists")

type outputExistsError struct {
	path string
}

func (e *outputExistsError) Error() string {
	return fmt.Sprintf("output ZIP already exists: %s", e.path)
}

func (e *outputExistsError) Is(target error) bool {
	return target == ErrOutputExists
}

func run(pdfPath string) error {
	zipPath, err := convertToZip(pdfPath, filepath.Dir(pdfPath), os.Stdout, true)
	if err != nil {
		return err
	}

	fmt.Println("Done:", filepath.Base(zipPath))
	return nil
}

func convertToZip(pdfPath, outDir string, progress io.Writer, overwrite bool) (string, error) {
	// 拡張子チェック（大文字小文字不問）
	if !strings.EqualFold(filepath.Ext(pdfPath), ".pdf") {
		return "", fmt.Errorf("not a PDF file: %s", filepath.Base(pdfPath))
	}

	// ファイル存在チェック
	if _, err := os.Stat(pdfPath); err != nil {
		return "", fmt.Errorf("cannot access file %s: %w", filepath.Base(pdfPath), err)
	}

	if info, err := os.Stat(outDir); err != nil {
		return "", fmt.Errorf("cannot access output directory: %s", outDir)
	} else if !info.IsDir() {
		return "", fmt.Errorf("output path is not a directory: %s", outDir)
	}

	// 出力先ディレクトリの書き込み権限チェック（変換前に確認）
	tmp, err := os.CreateTemp(outDir, ".pdf2png_check_*")
	if err != nil {
		return "", fmt.Errorf("cannot write to directory: %s", outDir)
	}
	tmp.Close()
	os.Remove(tmp.Name())

	// ZIP ファイルパスを生成
	base := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	zipPath := filepath.Join(outDir, base+".zip")
	if !overwrite {
		if _, err := os.Stat(zipPath); err == nil {
			return zipPath, &outputExistsError{path: zipPath}
		} else if !os.IsNotExist(err) {
			return zipPath, fmt.Errorf("cannot access output ZIP: %w", err)
		}
	}

	// 中間 PNG 用一時ディレクトリ（関数終了時に削除）
	tmpDir, err := os.MkdirTemp("", "pdf2png_*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// PDF → PNG 変換
	pngFiles, err := converter.Convert(pdfPath, tmpDir, progress)
	if err != nil {
		return "", err
	}

	// PNG → ZIP 圧縮
	if err := archiver.Archive(pngFiles, zipPath); err != nil {
		return zipPath, fmt.Errorf("failed to create ZIP: %w", err)
	}

	return zipPath, nil
}

// pause はユーザーの Enter キー入力を待つ（ドラッグ&ドロップ利用者向け）。
func pause() {
	fmt.Fprint(os.Stderr, "Press Enter to exit...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}
