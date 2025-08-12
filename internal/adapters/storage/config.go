package storage

import "product-catalog/internal/config"

type Config struct {
	Endpoint       string
	AccessKey      string
	SecretKey      string
	UseSSL         bool
	Bucket         string
	PublicEndpoint string
	Region         string
}

func NewConfig(c *config.Config) *Config {
	sc := c.Storage
	return &Config{
		Endpoint:  sc.Endpoint,
		AccessKey: sc.AccessKey,
		SecretKey: sc.SecretKey,
		UseSSL:    sc.UseSSL,
		Bucket:    sc.Bucket,
	}
}
