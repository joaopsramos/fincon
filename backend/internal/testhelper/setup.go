package testhelper

import (
	"log"
	"os"
	"path"
	"sync"

	"github.com/joaopsramos/fincon/internal/config"
)

var Setup = sync.OnceFunc(setupOnce)

func setupOnce() {
	err := os.Setenv("APP_ENV", "test")
	if err != nil {
		log.Fatal(err)
	}

	config.Load(path.Join("..", ".."))
}
