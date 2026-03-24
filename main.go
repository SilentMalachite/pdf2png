package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hiro/pdf2png/internal/archiver"
	"github.com/hiro/pdf2png/internal/converter"
)

func main() {
	// 引数チェック: usage エラーは pause なしで終了（ドラッグ&ドロップではなく誤操作）
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: pdf2png <file.pdf>")
		os.Exit(1)
	}

	if err := run(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		pauseOnError() // ドラッグ&ドロップ利用者がエラーを確認できるよう一時停止
		os.Exit(1)
	}
}

func run(pdfPath string) error {

	// 拡張子チェック（大文字小文字不問）
	if !strings.EqualFold(filepath.Ext(pdfPath), ".pdf") {
		return fmt.Errorf("not a PDF file: %s", filepath.Base(pdfPath))
	}

	// ファイル存在チェック
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filepath.Base(pdfPath))
	}

	// 出力先ディレクトリの書き込み権限チェック（変換前に確認）
	outDir := filepath.Dir(pdfPath)
	tmp, err := os.CreateTemp(outDir, ".pdf2png_check_*")
	if err != nil {
		return fmt.Errorf("cannot write to directory: %s", outDir)
	}
	tmp.Close()
	os.Remove(tmp.Name())

	// 中間 PNG 用一時ディレクトリ（関数終了時に削除）
	tmpDir, err := os.MkdirTemp("", "pdf2png_*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// PDF → PNG 変換（os.Stdout に進捗を出力）
	pngFiles, err := converter.Convert(pdfPath, tmpDir, os.Stdout)
	if err != nil {
		return err
	}

	// ZIP ファイルパスを生成（PDF と同じディレクトリ）
	base := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	zipPath := filepath.Join(outDir, base+".zip")

	// PNG → ZIP 圧縮
	if err := archiver.Archive(pngFiles, zipPath); err != nil {
		return fmt.Errorf("failed to create ZIP: %w", err)
	}

	fmt.Println("Done:", filepath.Base(zipPath))
	return nil
}

// pauseOnError はエラー時にユーザーの確認を待つ（ドラッグ&ドロップ利用者向け）。
func pauseOnError() {
	fmt.Fprint(os.Stderr, "Press Enter to exit...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}
