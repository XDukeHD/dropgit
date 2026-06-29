package database

import (
	"time"

	"database/sql"
)

func HasAnyBackupSession() (bool, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM backup_sessions WHERE successful_projects > 0").Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func CreateSession(sessionDate, startTime string) (int64, error) {
	stmt, err := DB.Prepare("INSERT INTO backup_sessions (session_date, start_time, status) VALUES (?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(sessionDate, startTime, "in_progress")
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateSession(id int64, endTime string, total, success, failed, size int, sha256, status string) error {
	stmt, err := DB.Prepare("UPDATE backup_sessions SET end_time = ?, total_projects = ?, successful_projects = ?, failed_projects = ?, total_size_bytes = ?, sha256 = ?, status = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(endTime, total, success, failed, size, sha256, status, id)
	return err
}

func CreateProjectBackup(sessionID int64, name, path, backupPath string, count, size int, sha256, status, errMsg string) (int64, error) {
	stmt, err := DB.Prepare("INSERT INTO backup_projects (session_id, project_name, project_path, backup_file_path, file_count, total_size_bytes, sha256, status, error_message) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(sessionID, name, path, backupPath, count, size, sha256, status, errMsg)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func RecordFileChange(sessionID int64, projectName, filePath, changeType, prevSha, currSha string, size int64) error {
	stmt, err := DB.Prepare("INSERT INTO file_changes (session_id, project_name, file_path, change_type, previous_sha256, current_sha256, file_size, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(sessionID, projectName, filePath, changeType, prevSha, currSha, size, time.Now().UTC().Format(time.RFC3339))
	return err
}

func GetPreviousFileHash(projectName, filePath string) (string, error) {
	var hash string
	err := DB.QueryRow("SELECT current_sha256 FROM file_changes WHERE project_name = ? AND file_path = ? ORDER BY id DESC LIMIT 1", projectName, filePath).Scan(&hash)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func GetLatestBackupEndTime() (time.Time, error) {
	var endTime sql.NullString
	err := DB.QueryRow("SELECT end_time FROM backup_sessions WHERE end_time IS NOT NULL AND end_time != '' ORDER BY id DESC LIMIT 1").Scan(&endTime)
	if err != nil {
		return time.Time{}, err
	}

	parsed, parseErr := time.Parse(time.RFC3339, endTime.String)
	if parseErr != nil {
		return time.Time{}, parseErr
	}

	return parsed, nil
}
