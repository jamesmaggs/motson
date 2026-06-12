package store_test

import (
	"testing"

	"github.com/jamesmaggs/motson/internal/fixtures"
	"github.com/jamesmaggs/motson/internal/store"
	"github.com/jamesmaggs/motson/internal/store/storetest"
)

func TestMemoryStoreContract(t *testing.T) {
	storetest.Run(t, func(t *testing.T) fixtures.Store {
		return store.NewMemory()
	})
}
