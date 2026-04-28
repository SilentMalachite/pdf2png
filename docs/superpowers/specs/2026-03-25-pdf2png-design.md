# pdf2png Design Spec

**Date:** 2026-03-25
**Status:** Approved

## Overview

Go製GUI/CLIツール。GUIではPDFファイルと出力先フォルダを選び、ページごとにPNG画像（300DPI）を生成してZIPファイルにまとめて出力する。CLIではPDFファイルパスを引数で受け取り、従来通りPDFと同じフォルダにZIPを出力する。社内利用向け単一バイナリ配布。

## Requirements

- 引数なし起動ではGUIを開く
- CLIではPDFファイルパスをコマンドライン引数（`os.Args[1]`）で受け取る
- GUIではPDFファイルと出力先フォルダを選択できる
- ページごとにPNG画像を生成（300DPI）
- PNG群をZIPに圧縮して出力
- ZIPファイル名はPDFファイル名から自動生成（例: `報告書2024.pdf` → `報告書2024.zip`）
- GUIのZIP出力先は選択した出力先フォルダ
- CLIのZIP出力先はPDFと同じディレクトリ
- GUIでは同名ZIPが存在する場合に上書き確認する
- 中間生成したPNGは最後に削除（一時ディレクトリ使用）
- macOS・Windows両対応
- 単一バイナリで配布
- GUIでは進捗ログ、完了ダイアログ、エラーダイアログを表示する
- CLIではエラー時に `Press Enter to exit` で一時停止する
- 成功時はそのまま終了（exit code 0）
- エラー時は exit code 1

## Library

**go-fitz** (`github.com/gen2brain/go-fitz`) — MuPDFのGoバインディング

- レンダリング品質: 業界標準レベル
- 静的リンクにより単一バイナリ実現（MuPDF静的ライブラリが必要）
- CGo使用
- ライセンス: AGPL（社内利用のため問題なし）
- DPI指定: `scale = 300.0 / 72.0 ≈ 4.167` のスケール係数で指定

GUIは **Fyne** (`fyne.io/fyne/v2`) を使用する。ZIP圧縮は標準ライブラリ `archive/zip` を使用。

## Project Structure

```
pdf2png/
├── main.go
├── gui.go              # Fyne GUI
├── main_test.go        # 出力先指定・上書き確認境界のテスト
├── internal/
│   ├── converter/
│   │   └── converter.go    # PDF→PNG変換（go-fitz）
│   └── archiver/
│       └── archiver.go     # PNG群→ZIP圧縮
├── .github/
│   └── workflows/
│       └── build.yml       # GitHub Actions CI
└── go.mod
```

## Data Flow

### GUI

```
[引数なし起動]
        ↓
  Fyne GUI を表示
        ↓
  PDFファイル選択
  - .pdf以外は変換開始時にエラー
  - PDF選択時、出力先フォルダの初期値はPDFと同じディレクトリ
        ↓
  出力先フォルダ選択（任意で変更）
        ↓
  変換開始
  - PDF/出力先未選択: エラーダイアログ
  - 出力先ディレクトリへの書き込み権限確認
  - 同名ZIPが存在: 上書き確認ダイアログ
        ↓
  変換中は操作ボタンを無効化し、進捗ログを表示
        ↓
  完了時はZIPパスをログとダイアログに表示
```

### CLI

```
[os.Args[1]: PDFファイルパス]
        ↓
  引数バリデーション
  - 引数なし: GUI起動
  - 引数複数: usage表示してexit(1)
  - ファイル不存在: エラーメッセージ → pause → exit(1)
  - .pdf以外の拡張子（大文字小文字不問）: エラー → pause → exit(1)
  - 出力ディレクトリへの書き込み権限確認（変換開始前）: エラー → pause → exit(1)
        ↓
  一時ディレクトリ作成 (os.MkdirTemp) → defer で削除
        ↓
  PDF open → 0ページ: "PDF has no pages" エラー → pause → exit(1)
  ページごとにPNG生成
  - ファイル名: page_001.png, page_002.png ... (総ページ数に応じて動的ゼロ埋め)
  - 進捗表示: "Converting page 1/10..."
        ↓
  PNG群をZIPに圧縮 (archive/zip)
  - 出力先: PDFと同じディレクトリ
  - 同名ZIPが存在する場合: 上書き（CLIのみ警告なし）
        ↓
  完了メッセージ表示 "Done: 報告書2024.zip"
  exit(0)
```

## Error Cases

| ケース | メッセージ例 | 動作 |
|--------|-------------|------|
| 引数なし | なし | GUI起動 |
| 引数複数 | `Usage: pdf2png <file.pdf>` | exit(1) |
| GUIでPDF未選択 | `PDFファイルを選択してください` | エラーダイアログ |
| GUIで出力先未選択 | `出力先フォルダを選択してください` | エラーダイアログ |
| GUIで同名ZIPあり | `... は既に存在します。上書きしますか？` | 確認ダイアログ |
| ファイル不存在 | `Error: cannot access file foo.pdf: ...` | pause → exit(1) |
| 出力先ディレクトリ不存在 | `Error: cannot access output directory: /path` | pause → exit(1) |
| .pdf以外の拡張子 | `Error: not a PDF file: foo.txt` | pause → exit(1) |
| 書き込み権限なし | `Error: cannot write to directory: /path` | pause → exit(1) |
| パスワード保護PDF | `Error: PDF is password-protected` | pause → exit(1) |
| 0ページPDF | `Error: PDF has no pages` | pause → exit(1) |
| 壊れたPDF | `Error: failed to open PDF: <detail>` | pause → exit(1) |
| 変換失敗（ページ単位） | `Error: failed to convert page N: <detail>` | pause → exit(1) |

**拡張子チェック**: `strings.EqualFold(ext, ".pdf")` で大文字小文字不問

**ゼロ埋め桁数**: `fmt.Sprintf("page_%0*d.png", len(strconv.Itoa(totalPages)), i+1)`

## Build & CI

GitHub ActionsでmacOS・Windowsネイティブビルド。

### 戦略

```yaml
strategy:
  matrix:
    include:
      - os: macos-13        # Intel (amd64) runner
        goarch: amd64
        artifact: pdf2png-darwin-amd64
      - os: macos-latest    # Apple Silicon (arm64) runner
        goarch: arm64
        artifact: pdf2png-darwin-arm64
      - os: windows-latest
        goarch: amd64
        artifact: pdf2png-windows-amd64.exe
```

### MuPDF静的ライブラリのプロビジョニング

**Embedded MuPDFモード（正規アプローチ）**を使用する。go-fitzは `CGO_ENABLED=1` でビルドするとMuPDFをソースからビルドしてバイナリに静的リンクする。追加のMuPDFインストールは不要。`go.mod` でgo-fitzのバージョンを固定する（`v0.13.0` 以降）。

**Cコンパイラの確保（必須）**:
- **macOS**: Xcodeコマンドラインツールが `macos-13` / `macos-latest` ランナーに標準搭載。追加設定不要。
- **Windows**: `windows-latest` ランナーにはMinGWが含まれないため、`msys2/setup-msys2` アクションでMinGW-w64 GCCを導入する：
  ```yaml
  - uses: msys2/setup-msys2@v2
    with:
      msystem: MINGW64
      install: mingw-w64-x86_64-gcc
  ```

### Releaseへのアップロード

タグプッシュ時に3バイナリをGitHub Releasesにアップロード。
