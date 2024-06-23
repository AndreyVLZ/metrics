package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	publicKeyType  = "PUBLIC KEY"
	privateKeyType = "PRIVATE KEY"
)

var errDecode = errors.New("decode pkBytes")

type Key struct {
	private *rsa.PrivateKey
	public  *rsa.PublicKey
}

func New(bitsSize int) (Key, error) {
	private, err := rsa.GenerateKey(rand.Reader, bitsSize)
	if err != nil {
		return Key{}, fmt.Errorf("generate key: %w", err)
	}

	return Key{
		private: private,
		public:  &private.PublicKey,
	}, nil
}

// WritePublicKeyTo Записывает в writer PEM-кодировку публичного ключа.
func (k *Key) WritePublicKeyTo(writer io.Writer) error {
	return pem.Encode(writer,
		&pem.Block{
			Type:  publicKeyType,
			Bytes: x509.MarshalPKCS1PublicKey(k.public),
		})
}

// WritePrivateKeyTo Записывает в writer PEM-кодировку приватного ключа.
func (k *Key) WritePrivateKeyTo(writer io.Writer) error {
	return pem.Encode(writer,
		&pem.Block{
			Type:  privateKeyType,
			Bytes: x509.MarshalPKCS1PrivateKey(k.private),
		})
}

// rsaPublicKey Читает публичный ключ rsa [rsa.PublicKey] из файла.
func RSAPublicKey(publicKeyPath string) (*rsa.PublicKey, error) {
	pkBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("open publicKey file: %w", err)
	}

	block, _ := pem.Decode(pkBytes)
	if block == nil || block.Type != publicKeyType {
		return nil, errDecode
	}

	return x509.ParsePKCS1PublicKey(block.Bytes)
}

// Encrypt Шифрует message публичным ключом.
func Encrypt(publicKey *rsa.PublicKey, message []byte) ([]byte, error) {
	cipher, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, message, nil)
	if err != nil {
		return nil, fmt.Errorf("encryptOAEP: %w", err)
	}

	return cipher, nil
}

// RSAPrivateKey Получает приватный ключ rsa [rsa.PrivateKey] из файла.
func RSAPrivateKey(privateKeyPath string) (*rsa.PrivateKey, error) {
	fmt.Printf("privateKeyPath: %v\n", privateKeyPath)
	pkBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("open publicKey file: %w", err)
	}

	block, _ := pem.Decode(pkBytes)
	if block == nil || block.Type != privateKeyType {
		return nil, errDecode
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// Decrypt Расшифровывает encryptMsg приватным ключом из pkBytes.
func Decrypt(privateKey *rsa.PrivateKey, encryptMsg []byte) ([]byte, error) {
	// расшифровываем encryptMsg
	cipher, err := rsa.DecryptOAEP(sha256.New(), nil, privateKey, encryptMsg, nil)
	if err != nil {
		return nil, fmt.Errorf("decryptOAEP: %w", err)
	}

	return cipher, nil
}
