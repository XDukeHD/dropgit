package scheduler

import (
	"github.com/XDukeHD/dropgit/internal/backup"
	"github.com/XDukeHD/dropgit/internal/config"
	"github.com/XDukeHD/dropgit/internal/logger"
	"github.com/robfig/cron/v3"
)

var c *cron.Cron

func Start() error {
	c = cron.New()

	var cronExpr string
	switch config.Cfg.ScheduleInterval {
	case "daily":
		cronExpr = "0 0 * * *"
	case "weekly":
		cronExpr = "0 0 * * 0"
	case "monthly":
		cronExpr = "0 0 1 * *"
	case "every_3h":
		cronExpr = "0 */3 * * *"
	case "every_6h":
		cronExpr = "0 */6 * * *"
	case "custom_cron":
		cronExpr = config.Cfg.CustomCronExpression
	default:
		logger.Log.Warnf("Unknown schedule_interval '%s', defaulting to daily", config.Cfg.ScheduleInterval)
		cronExpr = "0 0 * * *"
	}

	_, err := c.AddFunc(cronExpr, func() {
		backup.PerformBackup()
	})

	if err != nil {
		return err
	}

	c.Start()
	logger.Log.Infof("Scheduler started with expression: %s", cronExpr)
	return nil
}

func Stop() {
	if c != nil {
		c.Stop()
	}
}
