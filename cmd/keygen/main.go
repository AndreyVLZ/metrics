package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/AndreyVLZ/metrics/agent"
	"github.com/AndreyVLZ/metrics/pkg/crypto"
	mylog "github.com/AndreyVLZ/metrics/pkg/log"
	"github.com/AndreyVLZ/metrics/server"
)

func main() {
	// размер rsa ключа:
	// - 4096 max data len ~446
	// - 5120 max data len ~574
	size := 5120
	publicKeyPath := agent.CryproKeyPathDefault
	privateKeyPath := server.CryptoKeyPathDefault
	logger := mylog.New(mylog.SlogKey, "debug")

	key, err := crypto.New(size)
	if err != nil {
		logger.Error("crypto new", "error", err)

		return
	}

	if err := writeKey(publicKeyPath, privateKeyPath, key); err != nil {
		logger.Error("write key", "error", err)

		return
	}

	logger.Debug("keys",
		slog.Group("path",
			slog.String("public", publicKeyPath),
			slog.String("private", privateKeyPath),
		),
	)
}

func writeKey(publicKeyPath, privateKeyPath string, key crypto.Key) error {
	publicKeyFile, err := os.OpenFile(publicKeyPath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("open publicKey file: %w", err)
	}

	privateKeyFile, err := os.OpenFile(privateKeyPath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("open privateKey file: %w", err)
	}

	if err := key.WritePublicKeyTo(publicKeyFile); err != nil {
		return fmt.Errorf("write publicKey to file: %w", err)
	}

	if err := key.WritePrivateKeyTo(privateKeyFile); err != nil {
		return fmt.Errorf("write privateKey to file: %w", err)
	}

	return nil
}
