package vertex

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigURL(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected string
	}{
		{
			name:     "global location",
			config:   Config{ProjectId: "my-project", Location: "global", AppID: "my-app"},
			expected: "https://discoveryengine.googleapis.com/v1/projects/my-project/locations/global/collections/default_collection/engines/my-app/servingConfigs/default_search:search",
		},
		{
			name:     "us location",
			config:   Config{ProjectId: "my-project", Location: "us", AppID: "my-app"},
			expected: "https://us-discoveryengine.googleapis.com/v1/projects/my-project/locations/us/collections/default_collection/engines/my-app/servingConfigs/default_search:search",
		},
		{
			name:     "eu location",
			config:   Config{ProjectId: "my-project", Location: "eu", AppID: "my-app"},
			expected: "https://eu-discoveryengine.googleapis.com/v1/projects/my-project/locations/eu/collections/default_collection/engines/my-app/servingConfigs/default_search:search",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.URL()
			if got != tt.expected {
				t.Errorf("URL() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			content: `project_id: "123456"
location: "us"
app_id: "my-app"
`,
			wantErr: false,
		},
		{
			name: "missing project_id",
			content: `location: "us"
app_id: "my-app"
`,
			wantErr: true,
			errMsg:  "project_id must be set",
		},
		{
			name: "missing app_id",
			content: `project_id: "123456"
location: "us"
`,
			wantErr: true,
			errMsg:  "app_id must be set",
		},
		{
			name: "missing location",
			content: `project_id: "123456"
app_id: "my-app"
`,
			wantErr: true,
			errMsg:  "location must be set",
		},
		{
			name: "invalid location",
			content: `project_id: "123456"
location: "asia"
app_id: "my-app"
`,
			wantErr: true,
			errMsg:  "location must be one of",
		},
		{
			name:    "invalid yaml",
			content: `{{{invalid`,
			wantErr: true,
			errMsg:  "unmarshal config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "vertex.yaml")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			cfg, err := LoadConfig(tmpFile)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if cfg == nil {
					t.Error("expected config, got nil")
				}
			}
		})
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/vertex.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestLoadAppConfig(t *testing.T) {
	content := `project_id: "123456"
location: "global"
app_id: "my-app"
`
	tmpFile := filepath.Join(t.TempDir(), "vertex.yaml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("config from flag", func(t *testing.T) {
		app, err := LoadAppConfig(tmpFile, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if app.Config.ProjectId != "123456" {
			t.Errorf("ProjectId = %q, want %q", app.Config.ProjectId, "123456")
		}
		if app.IsDebug {
			t.Error("expected IsDebug to be false")
		}
	})

	t.Run("debug from flag", func(t *testing.T) {
		app, err := LoadAppConfig(tmpFile, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !app.IsDebug {
			t.Error("expected IsDebug to be true")
		}
	})

	t.Run("config from env", func(t *testing.T) {
		t.Setenv("VERTEX_CONFIG", tmpFile)
		app, err := LoadAppConfig("", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if app.Config.ProjectId != "123456" {
			t.Errorf("ProjectId = %q, want %q", app.Config.ProjectId, "123456")
		}
	})

	t.Run("debug from env", func(t *testing.T) {
		t.Setenv("VERTEX_DEBUG", "true")
		app, err := LoadAppConfig(tmpFile, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !app.IsDebug {
			t.Error("expected IsDebug to be true from VERTEX_DEBUG env")
		}
	})

	t.Run("no config provided", func(t *testing.T) {
		t.Setenv("VERTEX_CONFIG", "")
		_, err := LoadAppConfig("", false)
		if err == nil {
			t.Error("expected error when no config is provided")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
