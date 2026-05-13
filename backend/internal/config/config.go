package config

import (
	"fmt"
	"os"
	"strconv"

	"dario.cat/mergo"
	"github.com/willie68/schematic2/backend/internal/logging"
	"github.com/willie68/schematic2/backend/internal/services/health"
	"github.com/willie68/schematic2/backend/internal/services/shttp"
	"gopkg.in/yaml.v3"
)

const defaultConfigFile = "configs/service.yaml"

// Config contains runtime configuration for the backend service.
// It follows the go-micro style service.yaml structure.
type Config struct {
	SecretFile     string         `yaml:"secretfile"`
	HTTP           shttp.Config   `yaml:"http"`
	Healthcheck    health.Config  `yaml:"healthcheck"`
	Metrics        Metrics        `yaml:"metrics"`
	Logging        logging.Config `yaml:"logging"`
	Auth           Auth           `yaml:"auth"`
	Profiling      Profiling      `yaml:"profiling"`
	MongoDB        MongoDB        `yaml:"mongodb"`
	Repository     Repository     `yaml:"repository"`

	// ClientBasePath is the external base path prefix (e.g. /schematics2 for reverse-proxy).
	// Leave empty for direct container access. Set via CLIENT_BASE_PATH env var.
	ClientBasePath string `yaml:"clientBasePath,omitempty"`

	JWTSecret string `yaml:"jwtsecret,omitempty"`
	AdminUser string `yaml:"adminuser,omitempty"`
	AdminPass string `yaml:"adminpass,omitempty"`
}

type Metrics struct {
	Enable bool `yaml:"enable"`
}

type Auth struct {
	Type       string         `yaml:"type"`
	Properties map[string]any `yaml:"properties"`
}

type Profiling struct {
	Enable bool `yaml:"enable"`
}

type Repository struct {
	RepositoryPath      string `yaml:"repositoryPath"`
	ContainerMaxSizeMB  int64  `yaml:"containerMaxSizeMB"`
	CompressionType     string `yaml:"compressionType"` // "none", "gzip", "zstd"
}

type MongoDB struct {
	Hosts            []string `yaml:"hosts"`
	Host             string   `yaml:"host"`
	Port             int      `yaml:"port"`
	Username         string   `yaml:"username"`
	Password         string   `yaml:"password"`
	AuthDatabase     string   `yaml:"authDatabase"`
	AuthDB           string   `yaml:"authdb"`
	Database         string   `yaml:"database"`
	DirectConnection bool     `yaml:"directConnection"`
}

// LoadFromEnv loads the configuration from the default config file.
// Environment variables can be referenced in the YAML using ${VAR_NAME} syntax.
func LoadFromEnv() Config {
	cfg := defaultConfig()
	_ = cfg.loadFromFile(defaultConfigFile)
	cfg.finalize()
	return cfg
}

func defaultConfig() Config {
	return Config{
		ClientBasePath: "",
		HTTP: shttp.Config{
			Servicename: "go-micro",
			Port:        8080,
			SSLPort:     8443,
			ServiceURL:  "https://localhost:8443",
			DNSNames:    []string{"localhost"},
			IPAddresses: []string{"127.0.0.1"},
		},
		Healthcheck: health.Config{Period: 30, StartDelay: 3},
		Metrics:     Metrics{Enable: false},
		Logging:     logging.Config{Level: "debug"},
		Auth:        Auth{Properties: map[string]any{}},
		Profiling:   Profiling{Enable: false},
		MongoDB: MongoDB{
			Hosts:        []string{"127.0.0.1:27017"},
			Host:         "127.0.0.1",
			Port:         27017,
			AuthDatabase: "schematics",
			AuthDB:       "schematics",
			Database:     "schematics",
		},
		Repository: Repository{
			RepositoryPath:     "./repository",
			ContainerMaxSizeMB: 100,
		},
		JWTSecret: "change-me-in-production",
		AdminUser: "admin",
		AdminPass: "admin123",
	}
}

func (c *Config) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	expanded := os.ExpandEnv(string(data))
	if err := yaml.Unmarshal([]byte(expanded), c); err != nil {
		return err
	}
	if c.SecretFile != "" {
		if err := c.mergeSecretFile(c.SecretFile); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) mergeSecretFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var secret Config
	if err := yaml.Unmarshal(data, &secret); err != nil {
		return err
	}
	if err := mergo.Merge(c, secret, mergo.WithOverride); err != nil {
		return fmt.Errorf("merging secret file: %w", err)
	}
	return nil
}

func (c *Config) finalize() {
	if c.Auth.Properties == nil {
		c.Auth.Properties = map[string]any{}
	}
}

func (m MongoDB) GetHosts() []string {
	if len(m.Hosts) > 0 {
		return m.Hosts
	}
	if m.Host == "" {
		return nil
	}
	if m.Port > 0 {
		return []string{m.Host + ":" + strconv.Itoa(m.Port)}
	}
	return []string{m.Host}
}

func (m MongoDB) GetAuthDatabase() string {
	if m.AuthDatabase != "" {
		return m.AuthDatabase
	}
	return m.AuthDB
}
