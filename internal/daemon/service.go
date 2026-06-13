package daemon

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/XDukeHD/dropgit/internal/backup"
	"github.com/XDukeHD/dropgit/internal/config"
	"github.com/XDukeHD/dropgit/internal/database"
	"github.com/XDukeHD/dropgit/internal/logger"
	"github.com/XDukeHD/dropgit/internal/scheduler"
)

func Run() {
	if err := config.LoadConfig(); err != nil {
		logger.Log.Fatalf("Failed to load config: %v", err)
	}

	logger.InitLogger(config.Cfg.LogLevel)
	logger.Log.Info("DropGit daemon starting...")

	if err := database.InitDB(); err != nil {
		logger.Log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	hasBackup, err := database.HasAnyBackupSession()
	if err != nil {
		logger.Log.Errorf("Failed to check for existing backups: %v", err)
	} else if !hasBackup {
		logger.Log.Info("No previous backups found. Starting initial backup...")
		go backup.PerformBackup()
	}

	if err := scheduler.Start(); err != nil {
		logger.Log.Fatalf("Failed to start scheduler: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	for {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP:
			logger.Log.Info("Received SIGHUP, reloading configuration...")
			scheduler.Stop()
			if err := config.LoadConfig(); err != nil {
				logger.Log.Errorf("Error reloading config: %v", err)
			}
			logger.InitLogger(config.Cfg.LogLevel)
			if err := scheduler.Start(); err != nil {
				logger.Log.Errorf("Error restarting scheduler: %v", err)
			}
		case syscall.SIGINT, syscall.SIGTERM:
			logger.Log.Info("Received termination signal, shutting down gracefully...")
			scheduler.Stop()
			return
		}
	}
}

func RunOnce() {
	if err := config.LoadConfig(); err != nil {
		logger.Log.Fatalf("Failed to load config: %v", err)
	}
	logger.InitLogger(config.Cfg.LogLevel)

	if err := database.InitDB(); err != nil {
		logger.Log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	backup.PerformBackup()
}
