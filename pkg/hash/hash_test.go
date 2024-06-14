package hash

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSHAok(t *testing.T) {
	bKey := []byte("SECRET")
	bData := []byte("test string")

	hash, err := SHA256(bData, bKey)
	if err != nil {
		t.Errorf("sha256: %v\n", err)

		return
	}

	t.Run("valid", func(t *testing.T) {
		ok, err := ValidMAC(hex.EncodeToString(hash), bData, bKey)
		if !ok {
			t.Errorf("key: %v\n", err)

			return
		}

		if err != nil {
			t.Errorf("check MAC %v\n", err)
		}
	})

	t.Run("not valid", func(t *testing.T) {
		fakeKey := []byte("SECRET1")

		ok, err := ValidMAC(hex.EncodeToString(hash), bData, fakeKey)
		if ok {
			t.Errorf("is fake key: %v\n", err)

			return
		}

		if err != nil {
			t.Errorf("check MAC %v\n", err)
		}
	})

	t.Run("not TEST", func(t *testing.T) {
		key1 := []byte("SECRET-1")
		key2 := []byte("SECRET-2")
		data := []byte("custom strring")

		sData1, err := SHA256(data, key1)
		if err != nil {
			t.Errorf("build hash: %v\n", err)
		}

		sData2, err := SHA256(data, key2)
		if err != nil {
			t.Errorf("build hash: %v\n", err)
		}

		isValid, err := ValidMAC(hex.EncodeToString(sData2), data, key1)
		t.Logf("isValid: %v\n", isValid)
		if err != nil {
			t.Errorf("build hash: %v\n", err)
		}

		if isValid {
			t.Errorf("valid hash: %v\n", isValid)
		}

		assert.NotEqual(t,
			hex.EncodeToString(sData1),
			hex.EncodeToString(sData2),
		)
	})
}
