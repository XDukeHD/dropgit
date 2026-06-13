package backup

import (
	"os"
	"path/filepath"

	"github.com/XDukeHD/dropgit/internal/hasher"
	"github.com/XDukeHD/dropgit/internal/ignore"
	"github.com/XDukeHD/dropgit/internal/logger"
)

type FileInfo struct {
	Path   string
	Size   int64
	SHA256 string
}

func ScanProject(projectPath string) ([]FileInfo, error) {
	var files []FileInfo
	matcher := ignore.NewMatcher(projectPath)

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Log.Warnf("Error accessing path %s: %v", path, err)
			return nil
		}

		if path == projectPath {
			return nil
		}

		if matcher.IsIgnored(path, projectPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() {
			hash, hashErr := hasher.ComputeFileSHA256(path)
			if hashErr != nil {
				logger.Log.Warnf("Failed to hash file %s: %v", path, hashErr)
				hash = ""
			}

			files = append(files, FileInfo{
				Path:   path,
				Size:   info.Size(),
				SHA256: hash,
			})
		}

		return nil
	})

	return files, err
}
