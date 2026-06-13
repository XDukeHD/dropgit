package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/XDukeHD/dropgit/internal/config"
	"github.com/XDukeHD/dropgit/internal/database"
	"github.com/XDukeHD/dropgit/internal/hasher"
	"github.com/XDukeHD/dropgit/internal/logger"
	"github.com/XDukeHD/dropgit/pkg/utils"
)

func PerformBackup() {
	logger.Log.Info("Starting backup cycle")

	if err := utils.EnsureDir(config.Cfg.SourceDirectory); err != nil {
		logger.Log.Errorf("Source directory not accessible: %v", err)
		return
	}

	dateStr := utils.CurrentDateStr()
	destDir := filepath.Join(config.Cfg.BackupDestination, dateStr)
	if err := utils.EnsureDir(destDir); err != nil {
		logger.Log.Errorf("Could not create destination directory: %v", err)
		return
	}

	sessionID, err := database.CreateSession(dateStr, utils.CurrentDateTimeISO())
	if err != nil {
		logger.Log.Errorf("Could not create DB session: %v", err)
		return
	}

	entries, err := os.ReadDir(config.Cfg.SourceDirectory)
	if err != nil {
		logger.Log.Errorf("Could not read source directory: %v", err)
		database.UpdateSession(sessionID, utils.CurrentDateTimeISO(), 0, 0, 0, 0, "", "failed")
		return
	}

	var projects []string
	for _, entry := range entries {
		if entry.IsDir() {
			projects = append(projects, entry.Name())
		}
	}

	totalProjects := len(projects)
	if totalProjects == 0 {
		logger.Log.Info("No projects found to backup")
		database.UpdateSession(sessionID, utils.CurrentDateTimeISO(), 0, 0, 0, 0, "", "completed")
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, config.Cfg.ParallelBackups)

	type result struct {
		success bool
		size    int
	}

	results := make(chan result, totalProjects)

	for _, p := range projects {
		wg.Add(1)
		sem <- struct{}{}
		go func(projectName string) {
			defer wg.Done()
			defer func() { <-sem }()

			size, success := backupSingleProject(sessionID, projectName, destDir)
			results <- result{success: success, size: size}
		}(p)
	}

	wg.Wait()
	close(results)

	var successCount, failedCount, totalSize int
	for r := range results {
		if r.success {
			successCount++
		} else {
			failedCount++
		}
		totalSize += r.size
	}

	status := "completed"
	if failedCount > 0 {
		status = "completed_with_errors"
	}

	database.UpdateSession(sessionID, utils.CurrentDateTimeISO(), totalProjects, successCount, failedCount, totalSize, "", status)

	if config.Cfg.RetentionDays > 0 {
		CleanupOldBackups()
	}

	logger.Log.Infof("Backup cycle finished. Success: %d, Failed: %d", successCount, failedCount)
}

func backupSingleProject(sessionID int64, projectName, destDir string) (int, bool) {
	projectPath := filepath.Join(config.Cfg.SourceDirectory, projectName)
	archiveName := fmt.Sprintf("%s.tar.gz", projectName)
	archivePath := filepath.Join(destDir, archiveName)

	logger.Log.Infof("Scanning project: %s", projectName)
	files, err := ScanProject(projectPath)
	if err != nil {
		logger.Log.Errorf("Failed to scan %s: %v", projectName, err)
		database.CreateProjectBackup(sessionID, projectName, projectPath, "", 0, 0, "", "failed", err.Error())
		return 0, false
	}

	var totalSize int64
	for _, f := range files {
		totalSize += f.Size
	}

	var changelog *Changelog
	if config.Cfg.EnableChangelog {
		changelog, err = GenerateChangelog(sessionID, projectName, files)
		if err != nil {
			logger.Log.Warnf("Could not generate changelog for %s: %v", projectName, err)
		}
	}

	var needsBackup bool
	if changelog != nil && len(changelog.Changes) > 0 {
		needsBackup = true
	} else if changelog == nil {
		needsBackup = true
	} else {
		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			needsBackup = true
		}
	}

	if !needsBackup {
		logger.Log.Infof("No changes for project %s, skipping archive generation", projectName)
		database.CreateProjectBackup(sessionID, projectName, projectPath, archivePath, len(files), int(totalSize), "", "skipped", "")
		return 0, true
	}

	logger.Log.Infof("Archiving project: %s", projectName)
	if err := CreateArchive(files, projectPath, archivePath, config.Cfg.CompressionLevel); err != nil {
		logger.Log.Errorf("Failed to archive %s: %v", projectName, err)
		database.CreateProjectBackup(sessionID, projectName, projectPath, "", len(files), int(totalSize), "", "failed", err.Error())
		return 0, false
	}

	archiveHash := ""
	if config.Cfg.EnableSHA256Validation {
		archiveHash, err = hasher.ComputeFileSHA256(archivePath)
		if err != nil {
			logger.Log.Errorf("Failed to hash archive for %s: %v", projectName, err)
			database.CreateProjectBackup(sessionID, projectName, projectPath, archivePath, len(files), int(totalSize), "", "failed_validation", err.Error())
			return 0, false
		}
	}

	if changelog != nil {
		WriteChangelog(changelog, destDir)
	}

	archiveStat, _ := os.Stat(archivePath)
	archiveSize := 0
	if archiveStat != nil {
		archiveSize = int(archiveStat.Size())
	}

	database.CreateProjectBackup(sessionID, projectName, projectPath, archivePath, len(files), int(totalSize), archiveHash, "success", "")
	return archiveSize, true
}

func CleanupOldBackups() {
	entries, err := os.ReadDir(config.Cfg.BackupDestination)
	if err != nil {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -config.Cfg.RetentionDays)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		t, err := time.Parse("2006-01-02", entry.Name())
		if err == nil && t.Before(cutoff) {
			dirToRemove := filepath.Join(config.Cfg.BackupDestination, entry.Name())
			logger.Log.Infof("Cleaning up old backup: %s", dirToRemove)
			os.RemoveAll(dirToRemove)
		}
	}
}
