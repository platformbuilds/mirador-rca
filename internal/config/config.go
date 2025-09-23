package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config captures the minimal settings required to boot the RCA service.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Clients  ClientsConfig  `yaml:"clients"`
	Weaviate WeaviateConfig `yaml:"weaviate"`
	Logging  LoggingConfig  `yaml:"logging"`
	Rules    RulesConfig    `yaml:"rules"`
}

// ServerConfig controls gRPC listener behaviour.
type ServerConfig struct {
	Address         string        `yaml:"address"`
	GracefulTimeout time.Duration `yaml:"gracefulTimeout"`
}

// ClientsConfig groups integrations with Victoria* backends.
type ClientsConfig struct {
	Core CoreClientConfig `yaml:"core"`
}

// CoreClientConfig configures access to mirador-core data aggregation APIs.
type CoreClientConfig struct {
	BaseURL          string        `yaml:"baseURL"`
	MetricsPath      string        `yaml:"metricsPath"`
	LogsPath         string        `yaml:"logsPath"`
	TracesPath       string        `yaml:"tracesPath"`
	ServiceGraphPath string        `yaml:"serviceGraphPath"`
	Timeout          time.Duration `yaml:"timeout"`
}

// WeaviateConfig configures the similarity search cluster.
type WeaviateConfig struct {
	Endpoint string        `yaml:"endpoint"`
	APIKey   string        `yaml:"apiKey"`
	Timeout  time.Duration `yaml:"timeout"`
}

// LoggingConfig controls structured logging.
type LoggingConfig struct {
	Level string `yaml:"level"`
	JSON  bool   `yaml:"json"`
}

// RulesConfig controls rule-pack loading for the recommender.
type RulesConfig struct {
	Path string `yaml:"path"`
}

// Load initialises Config from a YAML file and optional environment overrides.
func Load(path string) (*Config, error) {
	if path == "" {
		path = os.Getenv("MIRADOR_RCA_CONFIG")
	}

	cfg := defaultConfig()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf("config file %s not found: %w", path, err)
			}
			return nil, fmt.Errorf("read config: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}

	applyEnvOverrides(&cfg)
	return &cfg, nil
}

func defaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Address:         ":50051",
			GracefulTimeout: 10 * time.Second,
		},
		Clients: ClientsConfig{
			Core: CoreClientConfig{
				MetricsPath:      "/api/v1/rca/metrics",
				LogsPath:         "/api/v1/rca/logs",
				TracesPath:       "/api/v1/rca/traces",
				ServiceGraphPath: "/api/v1/rca/service-graph",
				Timeout:          5 * time.Second,
			},
		},
		Weaviate: WeaviateConfig{Timeout: 5 * time.Second},
		Logging:  LoggingConfig{Level: "info", JSON: false},
		Rules:    RulesConfig{Path: "configs/rules/default.yaml"},
	}
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("MIRADOR_RCA_SERVER_ADDRESS"); v != "" {
		cfg.Server.Address = v
	}
	if v := os.Getenv("MIRADOR_CORE_BASE_URL"); v != "" {
		cfg.Clients.Core.BaseURL = v
	}
	if v := os.Getenv("MIRADOR_CORE_METRICS_PATH"); v != "" {
		cfg.Clients.Core.MetricsPath = v
	}
	if v := os.Getenv("MIRADOR_CORE_LOGS_PATH"); v != "" {
		cfg.Clients.Core.LogsPath = v
	}
	if v := os.Getenv("MIRADOR_CORE_TRACES_PATH"); v != "" {
		cfg.Clients.Core.TracesPath = v
	}
	if v := os.Getenv("MIRADOR_CORE_SERVICE_GRAPH_PATH"); v != "" {
		cfg.Clients.Core.ServiceGraphPath = v
	}
	if v := os.Getenv("MIRADOR_RCA_WEAVIATE_URL"); v != "" {
		cfg.Weaviate.Endpoint = v
	}
	if v := os.Getenv("MIRADOR_RCA_WEAVIATE_API_KEY"); v != "" {
		cfg.Weaviate.APIKey = v
	}
	if v := os.Getenv("MIRADOR_RCA_LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("MIRADOR_RCA_LOG_FORMAT"); v == "json" {
		cfg.Logging.JSON = true
	}
	if v := os.Getenv("MIRADOR_RCA_RULES_PATH"); v != "" {
		cfg.Rules.Path = v
	}
}
