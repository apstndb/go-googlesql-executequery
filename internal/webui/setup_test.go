package webui

import (
	"os"
	"testing"

	"github.com/apstndb/go-googlesql-executequery/cache"
)

func TestMain(m *testing.M) {
	if err := cache.Setup(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
