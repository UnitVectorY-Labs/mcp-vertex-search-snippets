package vertex

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ProjectId string `yaml:"project_id"`
	Location  string `yaml:"location"`
	AppID     string `yaml:"app_id"`
}

func (c *Config) URL() string {

	// Domain is based on location
	domain := "discoveryengine.googleapis.com"
	if c.Location == "us" {
		domain = "us-discoveryengine.googleapis.com"
	} else if c.Location == "eu" {
		domain = "eu-discoveryengine.googleapis.com"
	} else if c.Location == "global" {
		domain = "discoveryengine.googleapis.com"
	}

	return fmt.Sprintf(
		"https://%s/v1alpha/projects/%s/locations/%s/collections/default_collection/engines/%s/servingConfigs/default_search:search",
		domain,
		c.ProjectId,
		c.Location,
		c.AppID,
	)
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

	if c.ProjectId == "" {
		return nil, fmt.Errorf("project_id must be set in vertex.yaml")
	}

	if c.AppID == "" {
		return nil, fmt.Errorf("app_id must be set in vertex.yaml")
	}

	if c.Location == "" {
		return nil, fmt.Errorf("location must be set in vertex.yaml")
	}

	// The location can be 'global', 'us', or 'eu'
	if c.Location != "global" && c.Location != "us" && c.Location != "eu" {
		return nil, fmt.Errorf("location must be one of 'global', 'us', or 'eu'")
	}

	return &c, nil
}

type AppConfig struct {
	Config  *Config
	IsDebug bool
}

func LoadAppConfig(configFlag string, debugFlag bool) (*AppConfig, error) {
	// The path to the file
	cfgPath := ""
	if configFlag != "" {
		cfgPath = configFlag
	} else if env := os.Getenv("VERTEX_CONFIG"); env != "" {
		cfgPath = env
	} else {
		return nil, fmt.Errorf("configuration file must be set via --vertexConfig or VERTEX_CONFIG")
	}

	isDebug := debugFlag
	if !isDebug {
		if env := os.Getenv("VERTEX_DEBUG"); env != "" {
			isDebug, _ = strconv.ParseBool(env)
		}
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		return nil, err
	}

	return &AppConfig{Config: cfg, IsDebug: isDebug}, nil
}
