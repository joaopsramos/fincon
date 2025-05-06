package repository_test

import (
	"os"
	"testing"

	"github.com/joaopsramos/fincon/internal/testhelper"
)

func TestMain(m *testing.M) {
	testhelper.Setup()

	os.Exit(m.Run())
}
