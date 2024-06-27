package store

import (
	"testing"

	"github.com/AndreyVLZ/metrics/internal/store/filestore"
	"github.com/AndreyVLZ/metrics/internal/store/inmemory"
	"github.com/AndreyVLZ/metrics/internal/store/postgres"
	"github.com/AndreyVLZ/metrics/server/config"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type testCase struct {
		name      string
		storeName string
		cfg       config.StorageConfig
	}

	tc := []testCase{
		{
			name:      "memStore",
			storeName: inmemory.NameConst,
			cfg: config.StorageConfig{
				ConnDB:    "",
				StorePath: "",
			},
		},

		{
			storeName: filestore.NameConst,
			name:      "fileStore",
			cfg: config.StorageConfig{
				ConnDB:    "",
				StorePath: "-",
			},
		},

		{
			storeName: postgres.NameConst,
			name:      "postgresStore",
			cfg: config.StorageConfig{
				ConnDB:    "-",
				StorePath: "-",
			},
		},
	}

	for _, test := range tc {
		t.Run(test.name, func(t *testing.T) {
			store := New(test.cfg)
			assert.Equal(t, test.storeName, store.Name())
		})
	}
}
