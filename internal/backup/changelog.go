package backup

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/XDukeHD/dropgit/internal/database"
)

type ChangeType string

const (
	ChangeAdded    ChangeType = "added"
	ChangeModified ChangeType = "modified"
	ChangeDeleted  ChangeType = "deleted"
)

type FileChange struct {
	Path           string     `json:"path"`
	Type           ChangeType `json:"type"`
	PreviousSHA256 string     `json:"previous_sha256,omitempty"`
	CurrentSHA256  string     `json:"current_sha256,omitempty"`
}

type Changelog struct {
	ProjectName string       `json:"project_name"`
	Changes     []FileChange `json:"changes"`
}

func GenerateChangelog(sessionID int64, projectName string, currentFiles []FileInfo) (*Changelog, error) {
	cl := &Changelog{
		ProjectName: projectName,
		Changes:     []FileChange{},
	}

	for _, cf := range currentFiles {
		prevHash, err := database.GetPreviousFileHash(projectName, cf.Path)
		if err != nil {
			cl.Changes = append(cl.Changes, FileChange{
				Path:          cf.Path,
				Type:          ChangeAdded,
				CurrentSHA256: cf.SHA256,
			})
			database.RecordFileChange(sessionID, projectName, cf.Path, string(ChangeAdded), "", cf.SHA256, cf.Size)
		} else if prevHash != cf.SHA256 {
			cl.Changes = append(cl.Changes, FileChange{
				Path:           cf.Path,
				Type:           ChangeModified,
				PreviousSHA256: prevHash,
				CurrentSHA256:  cf.SHA256,
			})
			database.RecordFileChange(sessionID, projectName, cf.Path, string(ChangeModified), prevHash, cf.SHA256, cf.Size)
		} else {
			database.RecordFileChange(sessionID, projectName, cf.Path, "unchanged", prevHash, cf.SHA256, cf.Size)
		}
	}

	return cl, nil
}

func WriteChangelog(cl *Changelog, destDir string) error {
	if len(cl.Changes) == 0 {
		return nil
	}

	path := filepath.Join(destDir, ProjectFileStem(cl.ProjectName)+"_CHANGELOG.json")
	data, err := json.MarshalIndent(cl, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
