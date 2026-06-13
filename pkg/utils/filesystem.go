package utils

import (
	"os"
)

func EnsureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

func DirSize(path string) (int64, error) {
	var size int64
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, err
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if entry.IsDir() {
			subSize, _ := DirSize(path + "/" + entry.Name())
			size += subSize
		} else {
			size += info.Size()
		}
	}
	return size, nil
}
