# pdf2png

PDFファイルをページごとにPNG画像へ変換し、ZIPファイルにまとめて出力するCLIツールです。

## 特徴

- PDFの各ページを **300 DPI** のPNG画像に変換
- 生成したPNGを **ZIPファイル** にまとめて出力
- ZIPのファイル名はPDFから自動生成（例: `報告書2024.pdf` → `報告書2024.zip`）
- **macOS / Windows** 対応の単一バイナリで配布
- ファイルを **ドラッグ&ドロップ** するだけで使える

## ダウンロード

[Releases](../../releases) から対応するバイナリをダウンロードしてください。

| OS | ファイル |
|----|---------|
| macOS (Apple Silicon) | `pdf2png-darwin-arm64` |
| macOS (Intel) | `pdf2png-darwin-amd64` |
| Windows | `pdf2png-windows-amd64.exe` |

**macOSの場合:** ダウンロード後、実行権限を付与してください。

```bash
chmod +x pdf2png-darwin-arm64
```

初回起動時に「開発元を確認できません」と表示される場合は、**システム設定 → プライバシーとセキュリティ** から許可してください。

## 使い方

### ドラッグ&ドロップ

バイナリにPDFファイルをドラッグ&ドロップするだけで変換が始まります。

### コマンドライン

```bash
./pdf2png 報告書2024.pdf
```

変換が完了すると、PDFと同じフォルダに `報告書2024.zip` が生成されます。

```
Converting page 1/10...
Converting page 2/10...
...
Converting page 10/10...
Done: 報告書2024.zip
```

ZIP内のファイル構成:

```
報告書2024.zip
├── page_01.png
├── page_02.png
...
└── page_10.png
```

> ページ数に応じてファイル名のゼロ埋め桁数が自動調整されます（1ページなら `page_1.png`、10ページ以上なら `page_01.png` など）。

## エラー時の動作

エラーが発生した場合はメッセージを表示して一時停止します（ドラッグ&ドロップ利用時にウィンドウが閉じないよう）。

| 状況 | メッセージ |
|------|-----------|
| パスワード保護PDF | `PDF is password-protected` |
| ページ数0のPDF | `PDF has no pages` |
| 書き込み権限なし | `cannot write to directory: ...` |

## 開発者向け

### 必要なもの

- Go 1.26 以上
- Xcode コマンドラインツール（macOS）: `xcode-select --install`
- MinGW-w64（Windows）

### ビルド

```bash
git clone https://github.com/hiro/pdf2png.git
cd pdf2png
CGO_ENABLED=1 go build -o pdf2png .
```

### テスト

```bash
CGO_ENABLED=1 go test ./...
```

### リリース

`v` から始まるタグをpushすると GitHub Actions が自動でビルドし、Releases にバイナリをアップロードします。

```bash
git tag v1.0.0
git push origin v1.0.0
```

## ライセンス

本ツールは PDF レンダリングに [go-fitz](https://github.com/gen2brain/go-fitz)（[MuPDF](https://mupdf.com/)）を使用しています。

- **go-fitz / MuPDF**: AGPL-3.0
- 本ツール自体のライセンスは [LICENSE](LICENSE) を参照してください。

> **注意:** MuPDF は AGPL ライセンスです。社内利用・個人利用であれば問題ありませんが、商用配布を行う場合は別途ライセンスの確認が必要です。
