//go:build integration || e2e

package integration

import (
	"path/filepath"
	"runtime"
)

func projectRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "../../.."))
}

func MigrationsDir() string {
	return filepath.Join(projectRoot(), "migrations")
}

func RepositoryRoot() string {
	return projectRoot()
}
