package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Server   ServerConfig   `yaml:"server"`
	JWT      JWTConfig      `yaml:"jwt"`
	Database DatabaseConfig `yaml:"database"`
	Storage  StorageConfig  `yaml:"storage"`
}

type AppConfig struct {
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	Env      string `yaml:"env"`
	LogLevel string `yaml:"log_level"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type JWTConfig struct {
	TokenTTLSeconds int    `yaml:"token_ttl_seconds"`
	Secret          string `yaml:"-"`
}

type DatabaseConfig struct {
	Host    string `yaml:"host"`
	Port    string `yaml:"port"`
	User    string `yaml:"user"`
	Name    string `yaml:"name"`
	SSLMode string `yaml:"sslmode"`
	Pass    string `yaml:"-"`
}

type StorageConfig struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	UseSSL    bool   `yaml:"use_ssl"`
	Bucket    string `yaml:"bucket"`
}

var (
	cfg  *Config
	once sync.Once
)

func Load() *Config {
	once.Do(func() {
		err := godotenv.Load()
		if err != nil {
			log.Println("No .env file found or could not load it")
		}

		cfg = &Config{}

		f, err := os.Open("internal/config/config.yaml")
		if err != nil {
			log.Fatalf("failed to open config.yaml: %v", err)
		}
		defer f.Close()

		decoder := yaml.NewDecoder(f)
		if err = decoder.Decode(cfg); err != nil {
			log.Fatalf("failed to decode config.yaml: %v", err)
		}

		cfg.Storage.AccessKey = os.Getenv("MINIO_ACCESS_KEY")
		cfg.Storage.SecretKey = os.Getenv("MINIO_SECRET_KEY")

		cfg.JWT.Secret = os.Getenv("JWT_SECRET")
		cfg.Database.Pass = os.Getenv("DB_PASSWORD")

		if envEndpoint := os.Getenv("MINIO_ENDPOINT"); envEndpoint != "" {
			cfg.Storage.Endpoint = envEndpoint
		}

		if err = cfg.validate(); err != nil {
			log.Fatalf("configuration validation failed: %v", err)
		}
	})
	return cfg
}

func (c *Config) DSN() string {
	db := c.Database
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		db.User,
		db.Pass,
		db.Host,
		db.Port,
		db.Name,
		db.SSLMode,
	)
}

func (c *Config) validate() error {
	if c.Server.Port == "" {
		return errors.New("server.port is required")
	}
	if c.JWT.Secret == "" {
		return errors.New("JWT_SECRET is required")
	}
	if c.Database.Pass == "" {
		return errors.New("DB_PASSWORD is required")
	}
	if c.JWT.TokenTTLSeconds <= 0 {
		return errors.New("jwt.token_ttl_seconds must be positive")
	}
	if c.Storage.AccessKey == "" || c.Storage.SecretKey == "" {
		return errors.New("minio access key and secret key are required")
	}
	if c.Storage.Endpoint == "" {
		return errors.New("minio endpoint is required")
	}
	if c.Storage.Bucket == "" {
		return errors.New("minio bucket is required")
	}
	return nil
}

func (c *Config) TokenTTL() time.Duration {
	return time.Duration(c.JWT.TokenTTLSeconds) * time.Second
}
