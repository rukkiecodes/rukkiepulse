package auth

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type StoredConfig struct {
	Token string `yaml:"token"`
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".rukkie", "config.yaml"), nil
}

func saveToken(token string) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(StoredConfig{Token: token})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func loadStored() (*StoredConfig, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg StoredConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func clearToken() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte("token: \"\"\n"), 0600)
}
