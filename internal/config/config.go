package config

import (
	"fmt"

	"github.com/joeshaw/envdecode"
)

type AppConfig struct {
	Env        string `env:"ENV,default=development"`
	Port       string `env:"HTTP_PORT,default=8080"`
	AppName    string `env:"APP_NAME,default=ImobCheck API"`
	AppVersion string `env:"APP_VERSION,default=1.0.0"`

	// Database configuration
	DBHost     string `env:"DB_HOST,default=localhost"`
	DBPort     string `env:"DB_PORT,default=5432"`
	DBUser     string `env:"DB_USER,default=imobcheck"`
	DBPassword string `env:"DB_PASSWORD,default=imobcheck"`
	DBName     string `env:"DB_NAME,default=imobcheck"`
	DBSSLMode  string `env:"DB_SSL_MODE,default=disable"`

	// JWT configuration
	JWTSecret string `env:"JWT_SECRET,default=your-super-secret-jwt-key-change-this-in-production"`
	JWTExpiry int    `env:"JWT_EXPIRY,default=24"` // hours

	// Storage configuration
	StorageType      string `env:"STORAGE_TYPE,default=minio"`              // minio, local, or s3
	StorageEndpoint  string `env:"STORAGE_ENDPOINT,default=localhost:9000"` // MinIO endpoint or S3 endpoint
	StorageAccessKey string `env:"STORAGE_ACCESS_KEY,default=minioadmin"`
	StorageSecretKey string `env:"STORAGE_SECRET_KEY,default=minioadmin"`
	StorageBucket    string `env:"STORAGE_BUCKET,default=imobcheck-photos"`
	StorageUseSSL    bool   `env:"STORAGE_USE_SSL,default=false"`
	StorageBasePath  string `env:"STORAGE_BASE_PATH,default=./uploads"`                      // For local storage
	StoragePublicURL string `env:"STORAGE_PUBLIC_URL,default=http://localhost:8080/uploads"` // For local storage

	RabbitMQURL string `env:"RABBITMQ_URL,default=amqp://guest:guest@localhost:5672/"`
}

func InitConfig() (cfg *AppConfig, err error) {
	cfg = &AppConfig{}
	err = envdecode.Decode(cfg)
	return
}

// GetDatabaseDSN returns the database connection string
func (c *AppConfig) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}
