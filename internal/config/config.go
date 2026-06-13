package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	SourceDirectory        string `mapstructure:"source_directory"`
	BackupDestination      string `mapstructure:"backup_destination"`
	ScheduleInterval       string `mapstructure:"schedule_interval"`
	CustomCronExpression   string `mapstructure:"custom_cron_expression"`
	RetentionDays          int    `mapstructure:"retention_days"`
	CompressionLevel       int    `mapstructure:"compression_level"`
	ParallelBackups        int    `mapstructure:"parallel_backups"`
	EnableSHA256Validation bool   `mapstructure:"enable_sha256_validation"`
	EnableChangelog        bool   `mapstructure:"enable_changelog"`
	LogLevel               string `mapstructure:"log_level"`
}

var Cfg *Config

func LoadConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".config", "dropgit")
	configPath := filepath.Join(configDir, "config.yml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfig(configDir, configPath, homeDir); err != nil {
			return err
		}
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	setDefaults(homeDir)

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Warning: Syntax error in config file, using defaults: %v\n", err)
	}

	Cfg = &Config{}
	if err := viper.Unmarshal(Cfg); err != nil {
		return err
	}

	return nil
}

func setDefaults(homeDir string) {
	viper.SetDefault("source_directory", filepath.Join(homeDir, "Projects"))
	viper.SetDefault("backup_destination", filepath.Join(homeDir, "Documents", "DropGit", "backups"))
	viper.SetDefault("schedule_interval", "daily")
	viper.SetDefault("custom_cron_expression", "")
	viper.SetDefault("retention_days", 30)
	viper.SetDefault("compression_level", 6)
	viper.SetDefault("parallel_backups", 3)
	viper.SetDefault("enable_sha256_validation", true)
	viper.SetDefault("enable_changelog", true)
	viper.SetDefault("log_level", "info")
}

func createDefaultConfig(configDir, configPath, homeDir string) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	setDefaults(homeDir)
	return viper.WriteConfigAs(configPath)
}
