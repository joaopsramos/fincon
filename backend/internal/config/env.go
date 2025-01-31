package config

import (
	"os"
	"path"

	"github.com/joho/godotenv"
)

func LoadEnv(rootDir string) {
	var envFile string

	switch os.Getenv("APP_ENV") {
	case "prod":
		envFile = ".prod.env"
	case "test":
		envFile = ".test.env"
	default:
		envFile = ".env"
	}

	godotenv.Load(path.Join(rootDir, envFile))
}
