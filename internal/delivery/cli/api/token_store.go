package api

import (
	"errors"
	"os"
	"path/filepath"
)

type TokenStore interface {
	Get() (string, error)
	Set(string) error
	Clear() error
}

type FileTokenStore struct {
	path string
}

func NewFileTokenStore(path string) (*FileTokenStore, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, ".techui", "token")
	}
	return &FileTokenStore{path: path}, nil
}

func (s *FileTokenStore) Get() (string, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *FileTokenStore) Set(token string) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.path, []byte(token), 0o600)
}

func (s *FileTokenStore) Clear() error {
	if err := os.Remove(s.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
