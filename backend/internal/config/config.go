package config

import (
	"log"
	"os"
	"path"
	"sync"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv    string `env:"APP_ENV,required"`
	ApiURL    string `env:"APP_API_URL,required"`
	WebURL    string `env:"APP_WEB_URL,required"`
	SecretKey string `env:"SECRET_KEY,required"`

	// Honeybadger
	HoneybadgerAPIKey string `env:"HONEYBADGER_API_KEY"`

	// Mail Configuration
	MailDriver    string `env:"MAIL_DRIVER,required"`
	MailFromName  string `env:"MAIL_FROM_NAME,required"`
	MailFromEmail string `env:"MAIL_FROM_EMAIL,required"`

	AWSAccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`

	// Database Configuration
	Database struct {
		Host string `env:"POSTGRES_HOST,required"`
		Port string `env:"POSTGRES_PORT,required"`
		User string `env:"POSTGRES_USER,required"`
		Pass string `env:"POSTGRES_PASS,required"`
		Name string `env:"POSTGRES_DB,required"`
	}
}

var (
	cfg  = &Config{}
	once sync.Once
)

func Get() *Config {
	return cfg
}

func Load(rootDir string) {
	once.Do(func() {
		loadEnv(rootDir)

		if err := env.Parse(cfg); err != nil {
			log.Fatalf("failed to parse config: %v", err)
		}

		if err := env.Parse(&cfg.Database); err != nil {
			log.Fatalf("failed to parse database config: %v", err)
		}
	})
}

func loadEnv(rootDir string) {
	var envFile string

	switch os.Getenv("APP_ENV") {
	case "prod":
		envFile = ".prod.env"
	case "test":
		envFile = ".test.env"
	default:
		envFile = ".env"
	}

	err := godotenv.Load(path.Join(rootDir, envFile))
	if err != nil {
		panic(err)
	}
}
