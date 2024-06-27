package crypto

import (
	"crypto/rsa"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrypto(t *testing.T) {
	var (
		key        Key
		publicKey  *rsa.PublicKey
		privateKey *rsa.PrivateKey
		encMsg     []byte
	)

	msg := []byte("my message")
	publicKeyPath := "publicKey.pem"
	privateKeyPath := "privateKey.pem"

	publicKeyFile, err := os.OpenFile(publicKeyPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		t.Errorf("open file [%s]: %v\n", publicKeyPath, err)
	}

	privateKeyFile, err := os.OpenFile(privateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		t.Errorf("open file [%s]: %v\n", publicKeyPath, err)
	}

	t.Cleanup(func() {
		os.Remove(publicKeyPath)
		os.Remove(privateKeyPath)
	})

	t.Run("new key", func(t *testing.T) {
		var err error
		key, err = New(2048)
		if err != nil {
			t.Errorf("new key: %v\n", err)
		}
	})

	t.Run("write public key", func(t *testing.T) {
		if err := key.WritePublicKeyTo(publicKeyFile); err != nil {
			t.Errorf("write public key: %v\n", err)
		}
	})

	t.Run("write private key", func(t *testing.T) {
		if err := key.WritePrivateKeyTo(privateKeyFile); err != nil {
			t.Errorf("write private key: %v\n", err)
		}
	})

	t.Run("read public key", func(t *testing.T) {
		var err error

		publicKey, err = RSAPublicKey(publicKeyPath)
		if err != nil {
			t.Errorf("read public key: %v\n", err)
		}
	})

	t.Run("read private key", func(t *testing.T) {
		var err error

		privateKey, err = RSAPrivateKey(privateKeyPath)
		if err != nil {
			t.Errorf("read private key: %v\n", err)
		}
	})

	t.Run("encrypt", func(t *testing.T) {
		var err error

		encMsg, err = Encrypt(publicKey, msg)
		if err != nil {
			t.Errorf("encrypt: %v\n", err)
		}
	})

	t.Run("decrypt", func(t *testing.T) {
		var err error
		var resMsg []byte

		resMsg, err = Decrypt(privateKey, encMsg)
		if err != nil {
			t.Errorf("decrypt: %v\n", err)
		}

		assert.Equal(t, msg, resMsg)
	})
}
