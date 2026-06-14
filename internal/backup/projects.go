package backup

import (
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
		targets, walkErr := discoverProjectsInDir(sourceDirectory, rootPath)
		if walkErr != nil {
			continue
		}

		projects = append(projects, targets...)
	}

	return projects, nil
}

func discoverProjectsInDir(sourceDirectory, dirPath string) ([]ProjectTarget, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	foundProjects := make([]ProjectTarget, 0)
	hasRegularFiles := false
	hasProjectMarker := false

	for _, entry := range entries {
		if entry.IsDir() {
			if _, skip := discoverySkipDirs[entry.Name()]; skip {
				continue
			}

			childPath := filepath.Join(dirPath, entry.Name())
			childProjects, childErr := discoverProjectsInDir(sourceDirectory, childPath)
			if childErr != nil {
				continue
			}
			foundProjects = append(foundProjects, childProjects...)
			continue
		}

		hasRegularFiles = true
		if _, ok := projectMarkers[entry.Name()]; ok {
			hasProjectMarker = true
		}
	}

	includeCurrent := hasProjectMarker || (len(foundProjects) == 0 && hasRegularFiles)
	if includeCurrent {
		relPath, relErr := filepath.Rel(sourceDirectory, dirPath)
		if relErr != nil {
			return foundProjects, nil
		}

		foundProjects = append(foundProjects, ProjectTarget{
			Name: filepath.ToSlash(relPath),
			Path: dirPath,
		})
	}

	return foundProjects, nil
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
