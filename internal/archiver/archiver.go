package archiver

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// Archive creates a ZIP file at zipPath containing the given files.
// Each file is stored with only its base name.
// If zipPath already exists, it is overwritten.
func Archive(files []string, zipPath string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for _, src := range files {
		if err := addFile(w, src); err != nil {
			return err
		}
	}
	return nil
}

func addFile(w *zip.Writer, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	dst, err := w.Create(filepath.Base(src))
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, f)
	return err
}
