package tests

import (
	"os"
	"testing"

	"github.com/marcodenic/agentry/internal/env"
)

func TestMain(m *testing.M) {
	env.Load()
	os.Exit(m.Run())
}
