package config

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/AndreyVLZ/metrics/pkg/crypto"
	"github.com/AndreyVLZ/metrics/pkg/log"
)

var (
	AddressDefault        string        = "localhost:8080" // Значение по умолчанию для адреса эндпоинта HTTP-сервера.
	LogLevelDefault       string        = log.LevelErr     // Значение по умолчанию для уровня логирования.
	RateLimitDefault      int           = 3                // Значение по умолчания для количества одновременно исходящих запросов на сервер.
	PollIntervalDefault   time.Duration = 2 * time.Second  // Значение по умолчания для частоты опроса метрик из пакета runtime.
	ReportIntervalDefault time.Duration = 10 * time.Second // Значение по умолчания для частоты отправки метрик на сервер.
	// CryproKeyPathDefault  string        = "/tmp/public.pem" // Значение по умолчания для пути до файла с публичным ключом
)

// Config Структура кофигурации агента.
type Config struct {
	Addr           string
	PollInterval   time.Duration
	ReportInterval time.Duration
	RateLimit      int
	ConfigPath     string
	CryptoKeyPath  string
	Key            []byte
	PublicKey      *rsa.PublicKey
	LogLevel       string
}

func Default() *Config {
	return &Config{
		Addr:           AddressDefault,
		PollInterval:   PollIntervalDefault,
		ReportInterval: ReportIntervalDefault,
		RateLimit:      RateLimitDefault,
		LogLevel:       LogLevelDefault,
		//CryptoKeyPath:  CryproKeyPathDefault,
	}
}

// Опции для конфига.
type FuncOpt func(cfg *Config)

// New Возвращает новый конфиг из переданных опций.
func New(opts ...FuncOpt) (*Config, error) {
	var (
		err error
	)

	cfg := Default()

	for _, opt := range opts {
		opt(cfg)
	}

	// читаем публичный ключ из файла
	if cfg.CryptoKeyPath == "" {
		return cfg, nil
	}

	cfg.PublicKey, err = crypto.RSAPublicKey(cfg.CryptoKeyPath)
	if err != nil {
		return nil, fmt.Errorf("publicKey: %w", err)
	}

	return cfg, nil
}

// Установка адреса эндпоинта HTTP-сервера.
func SetAddr(addr string) FuncOpt {
	return func(cfg *Config) {
		cfg.Addr = addr
	}
}

// Установка частоты опроса метрик из пакета runtime.
func SetPollInterval(pollInterval time.Duration) FuncOpt {
	return func(cfg *Config) {
		cfg.PollInterval = pollInterval
	}
}

// Установка частоты отправки метрик на сервер.
func SetReportInterval(reportInterval time.Duration) FuncOpt {
	return func(cfg *Config) {
		cfg.ReportInterval = reportInterval
	}
}

// Установка количества одновременно исходящих запросов на сервер.
func SetRateLimit(rateLimit int) FuncOpt {
	return func(cfg *Config) {
		cfg.RateLimit = rateLimit
	}
}

// Установка ключа.
func SetKey(key string) FuncOpt {
	return func(cfg *Config) {
		cfg.Key = []byte(key)
	}
}

// Установка пути до файла с публичным ключом.
func SetCryptoKeyPath(cryptoKeyPath string) FuncOpt {
	return func(cfg *Config) {
		cfg.CryptoKeyPath = cryptoKeyPath
	}
}

// Установка уровня логирования.
func SetLogLevel(lvl string) FuncOpt {
	return func(cfg *Config) {
		cfg.LogLevel = lvl
	}
}

// Установка пути до файла конфигурации.
func SetConfigPath(configPath string) FuncOpt {
	return func(cfg *Config) {
		cfg.ConfigPath = configPath
	}
}
