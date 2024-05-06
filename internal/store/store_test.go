package store

import (
	"testing"

	"github.com/AndreyVLZ/metrics/internal/store/filestore"
	"github.com/AndreyVLZ/metrics/internal/store/memstore"
	"github.com/AndreyVLZ/metrics/internal/store/postgres"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type testCase struct {
		name      string
		storeName string
		cfg       Config
	}

	tc := []testCase{
		{
			name:      "memStore",
			storeName: memstore.NameConst,
			cfg: Config{
				ConnDB:    "",
				StorePath: "",
			},
		},

		{
			storeName: filestore.NameConst,
			name:      "fileStore",
			cfg: Config{
				ConnDB:    "",
				StorePath: "-",
			},
		},

		{
			storeName: postgres.NameConst,
			name:      "postgresStore",
			cfg: Config{
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
