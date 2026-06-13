package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dbDir := filepath.Join(homeDir, ".local", "share", "dropgit")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return err
	}

	dbPath := filepath.Join(dbDir, "backup.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	DB = db
	return migrate()
}

func migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS backup_sessions (
			id INTEGER PRIMARY KEY,
			session_date TEXT,
			start_time TEXT,
			end_time TEXT,
			total_projects INTEGER,
			successful_projects INTEGER,
			failed_projects INTEGER,
			total_size_bytes INTEGER,
			sha256 TEXT,
			status TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS backup_projects (
			id INTEGER PRIMARY KEY,
			session_id INTEGER,
			project_name TEXT,
			project_path TEXT,
			backup_file_path TEXT,
			file_count INTEGER,
			total_size_bytes INTEGER,
			sha256 TEXT,
			status TEXT,
			error_message TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS file_changes (
			id INTEGER PRIMARY KEY,
			session_id INTEGER,
			project_name TEXT,
			file_path TEXT,
			change_type TEXT,
			previous_sha256 TEXT,
			current_sha256 TEXT,
			file_size INTEGER,
			timestamp TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS ignore_patterns (
			id INTEGER PRIMARY KEY,
			pattern TEXT,
			is_default BOOLEAN,
			created_at TEXT
		);`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
