package config

import (
	"os"
	"path"
	"sync"

	"github.com/joho/godotenv"
)

var SecretKey = sync.OnceValue(secretKey)

func secretKey() []byte {
	return []byte(os.Getenv("SECRET_KEY"))
}

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

	err := godotenv.Load(path.Join(rootDir, envFile))
	if err != nil {
		panic(err)
	}
}
