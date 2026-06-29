package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/XDukeHD/dropgit/internal/backup"
	"github.com/XDukeHD/dropgit/internal/config"
	"github.com/XDukeHD/dropgit/internal/database"
	"github.com/XDukeHD/dropgit/internal/logger"
	"github.com/robfig/cron/v3"
)

var c *cron.Cron
var stopCh chan struct{}
var backupMu sync.Mutex

func cronExpression() string {
	switch config.Cfg.ScheduleInterval {
	case "daily":
		return "0 0 * * *"
	case "weekly":
		return "0 0 * * 0"
	case "monthly":
		return "0 0 1 * *"
	case "every_1h", "hourly":
		return "0 * * * *"
	case "every_3h":
		return "0 */3 * * *"
	case "every_6h":
		return "0 */6 * * *"
	case "custom_cron":
		return config.Cfg.CustomCronExpression
	default:
		logger.Log.Warnf("Unknown schedule_interval '%s', defaulting to daily", config.Cfg.ScheduleInterval)
		return "0 0 * * *"
	}
}

func runBackupIfDue() {
	backupMu.Lock()
	defer backupMu.Unlock()

	if !backupIsDue() {
		return
	}

	backup.PerformBackup()
}

func backupIsDue() bool {
	lastBackup, err := database.GetLatestBackupEndTime()
	if err != nil {
		return true
	}

	if lastBackup.IsZero() {
		return true
	}

	schedule, err := cron.ParseStandard(cronExpression())
	if err != nil {
		return true
	}

	nextRun := schedule.Next(lastBackup)
	return !time.Now().Before(nextRun)
}

func Start() error {
	c = cron.New()
	stopCh = make(chan struct{})

	cronExpr := cronExpression()
	if cronExpr == "" {
		return fmt.Errorf("invalid schedule expression")
	}

	_, err := c.AddFunc(cronExpr, func() {
		runBackupIfDue()
	})

	if err != nil {
		return err
	}

	c.Start()
	logger.Log.Infof("Scheduler started with expression: %s", cronExpr)

	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		runBackupIfDue()

		for {
			select {
			case <-ticker.C:
				runBackupIfDue()
			case <-stopCh:
				return
			}
		}
	}()

	return nil
}

func Stop() {
	if c != nil {
		c.Stop()
	}

	if stopCh != nil {
		close(stopCh)
		stopCh = nil
	}
}
