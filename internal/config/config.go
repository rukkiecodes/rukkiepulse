package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Project      string                 `yaml:"project"`
	Environments map[string]Environment `yaml:"environments"`
}

type Environment struct {
	Services []Service `yaml:"services"`
}

type Service struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	Type string `yaml:"type"` // REST | GRAPHQL
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("rukkie.yaml not found in current directory — create one to get started")
		}
		return nil, fmt.Errorf("failed to read rukkie.yaml: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid rukkie.yaml: %w", err)
	}

	return &cfg, nil
}

func (c *Config) GetServices(env string) ([]Service, error) {
	e, ok := c.Environments[env]
	if !ok {
		return nil, fmt.Errorf("environment %q not found in rukkie.yaml", env)
	}
	if len(e.Services) == 0 {
		return nil, fmt.Errorf("no services defined under environment %q", env)
	}
	return e.Services, nil
}
