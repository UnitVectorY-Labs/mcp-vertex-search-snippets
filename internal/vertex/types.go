package vertex

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	URL string `yaml:"url"` // Full servingConfigs:search URL
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if c.URL == "" {
		return nil, fmt.Errorf("url must be set in vertex.yaml")
	}
	return &c, nil
}

type AppConfig struct {
	ConfigDir string
	Config    *Config
	IsDebug   bool
}

func LoadAppConfig(configFlag string, debugFlag bool) (*AppConfig, error) {
	dir := ""
	if configFlag != "" {
		dir = configFlag
	} else if env := os.Getenv("VERTEX_CONFIG"); env != "" {
		dir = env
	} else {
		return nil, fmt.Errorf("configuration directory must be set via --vertexConfig or VERTEX_CONFIG")
	}

	isDebug := debugFlag
	if !isDebug {
		if env := os.Getenv("VERTEX_DEBUG"); env != "" {
			isDebug, _ = strconv.ParseBool(env)
		}
	}

	cfg, err := LoadConfig(filepath.Join(dir, "vertex.yaml"))
	if err != nil {
		return nil, err
	}

	return &AppConfig{ConfigDir: dir, Config: cfg, IsDebug: isDebug}, nil
}
