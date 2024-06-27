package config

import (
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/AndreyVLZ/metrics/pkg/crypto"
)

// Config Структура кофигурации агента.
type Config struct {
	Addr           string
	PollInterval   int
	ReportInterval int
	RateLimit      int
	Key            []byte
	CryptoKeyPath  string
	PublicKey      *rsa.PublicKey
}

// Опции для конфига.
type FuncOpt func(cfg *Config)

// New Возвращает новый конфиг из переданных опций.
func New(opts ...FuncOpt) (*Config, error) {
	var (
		cfg Config
		err error
	)

	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Addr == "" {
		return nil, errors.New("config address is empty")
	}

	if cfg.CryptoKeyPath == "" {
		return &cfg, nil
	}

	cfg.PublicKey, err = crypto.RSAPublicKey(cfg.CryptoKeyPath)
	if err != nil {
		return nil, fmt.Errorf("publicKey: %w", err)
	}

	return &cfg, nil
}

// Установка адреса эндпоинта HTTP-сервера.
func SetAddr(addr string) FuncOpt {
	return func(cfg *Config) {
		cfg.Addr = addr
	}
}

// Установка частоты опроса метрик из пакета runtime.
func SetPollInterval(pollInterval int) FuncOpt {
	return func(cfg *Config) {
		cfg.PollInterval = pollInterval
	}
}

// Установка частоты отправки метрик на сервер.
func SetReportInterval(reportInterval int) FuncOpt {
	return func(cfg *Config) {
		cfg.ReportInterval = reportInterval
	}
}

// Установка ключа.
func SetKey(key string) FuncOpt {
	return func(cfg *Config) {
		cfg.Key = []byte(key)
	}
}

// Установка количества одновременно исходящих запросов на сервер.
func SetRateLimit(rateLimit int) FuncOpt {
	return func(cfg *Config) {
		cfg.RateLimit = rateLimit
	}
}

// Установка пути до файла с публичным ключом.
func SetCryptoKeyPath(cryptoKeyPath string) FuncOpt {
	return func(cfg *Config) {
		cfg.CryptoKeyPath = cryptoKeyPath
	}
}
