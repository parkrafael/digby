package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type Config struct {
	AgentID     string `json:"agent_id"`
	WatchFolder string `json:"watch_folder"`
}

var configPath = filepath.Join(os.Getenv("HOME"), ".config", "digby", "config.json")

func Load() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Setup(input string) (*Config, error) {
	cfg := Config{
		AgentID:     uuid.New().String(),
		WatchFolder: input,
	}

	err := os.MkdirAll(filepath.Dir(configPath), 0755)
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
