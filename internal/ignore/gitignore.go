package ignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

var defaultIgnores = []string{
	".next", ".nuxt", "dist", "node_modules", "build", "target", "__pycache__", ".cache", ".git", ".idea", ".vscode", "coverage", ".env", ".venv", "venv", "env", "vendor", "bin", "obj", "out", "logs", "tmp", "temp", ".tmp", "*.log", "*.cache", "*.pyc", "*.pyo", "*.class", "*.o", "*.so", "*.dll", "*.exe",
}

type Matcher struct {
	patterns []string
}

func NewMatcher(projectPath string) *Matcher {
	m := &Matcher{patterns: make([]string, len(defaultIgnores))}
	copy(m.patterns, defaultIgnores)

	ignoreFile := filepath.Join(projectPath, ".dropgitignore")
	if file, err := os.Open(ignoreFile); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				m.patterns = append(m.patterns, line)
			}
		}
	}
	return m
}

func (m *Matcher) IsIgnored(path, projectPath string) bool {
	relPath, err := filepath.Rel(projectPath, path)
	if err != nil {
		return false
	}
	
	relPath = filepath.ToSlash(relPath)
	parts := strings.Split(relPath, "/")

	for _, pattern := range m.patterns {
		pattern = filepath.ToSlash(pattern)
		isDirPattern := strings.HasSuffix(pattern, "/")
		cleanPattern := strings.TrimSuffix(pattern, "/")

		for _, part := range parts {
			if matched, _ := filepath.Match(cleanPattern, part); matched {
				if !isDirPattern {
					return true
				} else {
					stat, err := os.Stat(path)
					if err == nil && stat.IsDir() {
						return true
					}
				}
			}
		}

		if matched, _ := filepath.Match(cleanPattern, relPath); matched {
			if !isDirPattern {
				return true
			} else {
				stat, err := os.Stat(path)
				if err == nil && stat.IsDir() {
					return true
				}
			}
		}
	}
	return false
}
