package store_test

import (
	"testing"

	"github.com/Jazzatola/motson/internal/fixtures"
	"github.com/Jazzatola/motson/internal/store"
	"github.com/Jazzatola/motson/internal/store/storetest"
)

func TestMemoryStoreContract(t *testing.T) {
	storetest.Run(t, func(t *testing.T) fixtures.Store {
		return store.NewMemory()
	})
}
