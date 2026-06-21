package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Database struct {
	Driver  string `json:"driver"` // sqlite | mysql
	SQLite  struct {
		Path string `json:"path"`
	} `json:"sqlite"`
	MySQL struct {
		DSN string `json:"dsn"`
	} `json:"mysql"`
	Timezone string `json:"timezone"` // e.g. Asia/Shanghai, default Local
}

type Server struct {
	Addr string `json:"addr"` // e.g. 127.0.0.1:8090
	CORS bool   `json:"cors"` // enable CORS for dev
}

type Config struct {
	Server   Server   `json:"server"`
	Database Database `json:"database"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

func validate(cfg *Config) error {
	switch cfg.Database.Driver {
	case "sqlite", "mysql", "":
	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}
	return nil
}

func applyDefaults(cfg *Config) {
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = "sqlite"
	}
	if cfg.Database.SQLite.Path == "" {
		cfg.Database.SQLite.Path = "./data/daily.db"
	}
	if cfg.Server.Addr == "" {
		cfg.Server.Addr = "127.0.0.1:8090"
	}
	if cfg.Database.Timezone == "" {
		cfg.Database.Timezone = "Local"
	}
}
