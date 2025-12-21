package config

import (
	"github.com/joeshaw/envdecode"
)

// AppConfig holds all configuration for the application
type AppConfig struct {
	// Server configuration
	Port string `env:"PORT,default=8080"`

	// Database configuration
	DatabaseURL string `env:"DATABASE_URL,required"`

	// Queue configuration
	RabbitMQURL string `env:"RABBITMQ_URL,default=amqp://guest:guest@localhost:5672/"`

	// Storage configuration
	StorageType     string `env:"STORAGE_TYPE,default=local"`
	StorageBucket   string `env:"STORAGE_BUCKET,default=nfce"`
	StorageBasePath string `env:"STORAGE_BASE_PATH,default=./storage"`

	// S3/MinIO configuration
	StorageEndpoint  string `env:"STORAGE_ENDPOINT"`
	StorageAccessKey string `env:"STORAGE_ACCESS_KEY"`
	StorageSecretKey string `env:"STORAGE_SECRET_KEY"`
	StorageUseSSL    bool   `env:"STORAGE_USE_SSL,default=true"`
	StoragePublicURL string `env:"STORAGE_PUBLIC_URL"`

	// JWT configuration
	JWTSecret string `env:"JWT_SECRET,required"`
	JWTExpiry int    `env:"JWT_EXPIRY,default=24"`

	// SEFAZ configuration
	SEFAZTimeout int `env:"SEFAZ_TIMEOUT,default=30"`

	// Logging
	LogLevel string `env:"LOG_LEVEL,default=info"`
}

// InitConfig initializes the application configuration from environment variables
func InitConfig() (*AppConfig, error) {
	var cfg AppConfig
	err := envdecode.Decode(&cfg)
	return &cfg, err
}
