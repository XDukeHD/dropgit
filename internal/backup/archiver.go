package backup

import (
	"fmt"
	"os"

	"github.com/mholt/archiver/v3"
)

func CreateArchive(files []FileInfo, projectPath, outputPath string, compressionLevel int) error {
	tarGz := archiver.NewTarGz()
	tarGz.CompressionLevel = compressionLevel

	var filePaths []string
	for _, f := range files {
		filePaths = append(filePaths, f.Path)
	}

	tempFile := outputPath + ".tmp.tar.gz"
	if err := tarGz.Archive(filePaths, tempFile); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to create archive: %w", err)
	}

	if err := os.Rename(tempFile, outputPath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename temp archive: %w", err)
	}

	return nil
}
