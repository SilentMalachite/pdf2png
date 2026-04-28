package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type guiProgressWriter struct {
	appendLog func(string)
}

func (w guiProgressWriter) Write(p []byte) (int, error) {
	w.appendLog(string(p))
	return len(p), nil
}

func runGUI() {
	a := app.New()
	w := a.NewWindow("pdf2png")
	w.Resize(fyne.NewSize(640, 420))

	var pdfPath string
	var outputDir string
	var converting bool

	pdfEntry := widget.NewEntry()
	pdfEntry.SetPlaceHolder("PDFファイル")
	pdfEntry.Disable()

	outputEntry := widget.NewEntry()
	outputEntry.SetPlaceHolder("出力先フォルダ")
	outputEntry.Disable()

	statusLabel := widget.NewLabel("PDFを選択してください")
	progress := widget.NewProgressBarInfinite()
	progress.Hide()

	logEntry := widget.NewMultiLineEntry()
	logEntry.SetMinRowsVisible(8)
	logEntry.Disable()

	var startButton *widget.Button
	var pdfButton *widget.Button
	var outputButton *widget.Button

	appendLog := func(text string) {
		fyne.Do(func() {
			logEntry.Append(text)
		})
	}

	updateControls := func() {
		if converting {
			pdfButton.Disable()
			outputButton.Disable()
			startButton.Disable()
			return
		}

		pdfButton.Enable()
		outputButton.Enable()
		if pdfPath != "" && outputDir != "" {
			startButton.Enable()
		} else {
			startButton.Disable()
		}
	}

	setBusy := func(busy bool) {
		converting = busy
		if busy {
			statusLabel.SetText("変換中...")
			progress.Show()
			progress.Start()
		} else {
			progress.Stop()
			progress.Hide()
		}
		updateControls()
	}

	startConversion := func(overwrite bool) {}
	startConversion = func(overwrite bool) {
		if converting {
			return
		}
		if strings.TrimSpace(pdfPath) == "" {
			dialog.ShowError(fmt.Errorf("PDFファイルを選択してください"), w)
			return
		}
		if strings.TrimSpace(outputDir) == "" {
			dialog.ShowError(fmt.Errorf("出力先フォルダを選択してください"), w)
			return
		}

		logEntry.SetText("")
		appendLog("変換を開始します\n")
		setBusy(true)

		go func() {
			zipPath, err := convertToZip(pdfPath, outputDir, guiProgressWriter{appendLog: appendLog}, overwrite)
			fyne.Do(func() {
				setBusy(false)

				if errors.Is(err, ErrOutputExists) {
					dialog.ShowConfirm(
						"上書き確認",
						fmt.Sprintf("%s は既に存在します。上書きしますか？", filepath.Base(zipPath)),
						func(ok bool) {
							if ok {
								startConversion(true)
								return
							}
							statusLabel.SetText("キャンセルしました")
							appendLog("上書きをキャンセルしました\n")
						},
						w,
					)
					return
				}

				if err != nil {
					statusLabel.SetText("エラー")
					appendLog("Error: " + err.Error() + "\n")
					dialog.ShowError(err, w)
					return
				}

				statusLabel.SetText("完了: " + filepath.Base(zipPath))
				appendLog("Done: " + zipPath + "\n")
				dialog.ShowInformation("完了", "ZIPを作成しました:\n"+zipPath, w)
			})
		}()
	}

	pdfButton = widget.NewButtonWithIcon("PDFを選択", theme.FileIcon(), func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			pdfPath = reader.URI().Path()
			pdfEntry.SetText(pdfPath)
			if outputDir == "" {
				outputDir = filepath.Dir(pdfPath)
				outputEntry.SetText(outputDir)
			}
			statusLabel.SetText("変換できます")
			updateControls()
		}, w)
		fd.SetFilter(&storage.ExtensionFileFilter{Extensions: []string{".pdf", ".PDF"}})
		fd.Show()
	})

	outputButton = widget.NewButtonWithIcon("出力先を選択", theme.FolderOpenIcon(), func() {
		fd := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uri == nil {
				return
			}

			outputDir = uri.Path()
			outputEntry.SetText(outputDir)
			statusLabel.SetText("変換できます")
			updateControls()
		}, w)
		fd.Show()
	})

	startButton = widget.NewButtonWithIcon("変換開始", theme.MediaPlayIcon(), func() {
		startConversion(false)
	})

	updateControls()

	fileRow := container.NewBorder(nil, nil, nil, pdfButton, pdfEntry)
	outputRow := container.NewBorder(nil, nil, nil, outputButton, outputEntry)
	actions := container.NewHBox(startButton)
	content := container.NewVBox(
		widget.NewLabelWithStyle("pdf2png", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("PDFファイル"),
		fileRow,
		widget.NewLabel("出力先フォルダ"),
		outputRow,
		actions,
		progress,
		statusLabel,
		logEntry,
	)

	w.SetContent(container.NewPadded(content))
	w.ShowAndRun()
}
