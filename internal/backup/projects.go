package backup

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type ProjectTarget struct {
	Name string
	Path string
}

var projectMarkers = map[string]struct{}{
	".git":                {},
	".dropgitignore":      {},
	"go.mod":              {},
	"package.json":        {},
	"pyproject.toml":      {},
	"requirements.txt":    {},
	"Pipfile":             {},
	"poetry.lock":         {},
	"Cargo.toml":          {},
	"pom.xml":             {},
	"build.gradle":        {},
	"build.gradle.kts":    {},
	"settings.gradle":     {},
	"settings.gradle.kts": {},
	"composer.json":       {},
	"Gemfile":             {},
	"mix.exs":             {},
	"Makefile":            {},
	"docker-compose.yml":  {},
	"docker-compose.yaml": {},
	"Dockerfile":          {},
	"README.md":           {},
	"main.go":             {},
	"index.html":          {},
	"index.js":            {},
	"manage.py":           {},
}

var discoverySkipDirs = map[string]struct{}{
	".git":         {},
	"node_modules": {},
	"dist":         {},
	"build":        {},
	"target":       {},
	"vendor":       {},
	"__pycache__":  {},
	".idea":        {},
	".vscode":      {},
	"tmp":          {},
	"temp":         {},
}

var errFound = errors.New("found")

func hasAnyFiles(dirPath string) bool {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if path != dirPath {
				if _, skip := discoverySkipDirs[info.Name()]; skip {
					return filepath.SkipDir
				}
			}
			return nil
		}
		return errFound
	})
	return err == errFound
}

func DiscoverProjects(sourceDirectory string) ([]ProjectTarget, error) {
	entries, err := os.ReadDir(sourceDirectory)
	if err != nil {
		return nil, err
	}

	projects := make([]ProjectTarget, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		rootPath := filepath.Join(sourceDirectory, entry.Name())
		targets, hasMarkerInTree := discoverProjectsInDir(sourceDirectory, rootPath)
		
		if !hasMarkerInTree && len(targets) == 0 {
			if hasAnyFiles(rootPath) {
				relPath, _ := filepath.Rel(sourceDirectory, rootPath)
				projects = append(projects, ProjectTarget{
					Name: filepath.ToSlash(relPath),
					Path: rootPath,
				})
			}
		} else {
			projects = append(projects, targets...)
		}
	}

	return projects, nil
}

type childResult struct {
	path      string
	projects  []ProjectTarget
	hadMarker bool
}

func discoverProjectsInDir(sourceDirectory, dirPath string) ([]ProjectTarget, bool) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, false
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			if _, ok := projectMarkers[entry.Name()]; ok {
				relPath, _ := filepath.Rel(sourceDirectory, dirPath)
				return []ProjectTarget{{
					Name: filepath.ToSlash(relPath),
					Path: dirPath,
				}}, true
			}
		}
	}

	var foundProjects []ProjectTarget
	anyChildHadMarker := false
	var results []childResult

	for _, entry := range entries {
		if entry.IsDir() {
			if _, skip := discoverySkipDirs[entry.Name()]; skip {
				continue
			}
			childPath := filepath.Join(dirPath, entry.Name())
			cp, hadMarker := discoverProjectsInDir(sourceDirectory, childPath)
			results = append(results, childResult{childPath, cp, hadMarker})
			if hadMarker {
				anyChildHadMarker = true
			}
		}
	}

	if anyChildHadMarker {
		for _, r := range results {
			if r.hadMarker {
				foundProjects = append(foundProjects, r.projects...)
			} else {
				if hasAnyFiles(r.path) {
					relPath, _ := filepath.Rel(sourceDirectory, r.path)
					foundProjects = append(foundProjects, ProjectTarget{
						Name: filepath.ToSlash(relPath),
						Path: r.path,
					})
				}
			}
		}
		return foundProjects, true
	}

	return nil, false
}

func ProjectArchiveName(projectName string) string {
	safeName := strings.ReplaceAll(projectName, "\\", "/")
	safeName = strings.ReplaceAll(safeName, "/", "__")
	safeName = strings.ReplaceAll(safeName, " ", "_")
	return safeName + ".tar.gz"
}

func ProjectFileStem(projectName string) string {
	safeName := strings.ReplaceAll(projectName, "\\", "/")
	safeName = strings.ReplaceAll(safeName, "/", "__")
	safeName = strings.ReplaceAll(safeName, " ", "_")
	return safeName
}
