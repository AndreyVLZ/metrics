package config

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/AndreyVLZ/metrics/pkg/crypto"
	"github.com/AndreyVLZ/metrics/pkg/log"
)

const (
	AddressDefault       string        = "localhost:8080"       // Значение по умолчанию для адреса эндпоинта HTTP-сервера.
	StorePathDefault     string        = "/tmp/metrics-db.json" // Значение по умолчанию для имени файла, куда сохраняются текущие значения.
	StoreIntervalDefault time.Duration = 300 * time.Second      // Значение по умолчанию для интервала времени в секундах, по истечении которого текущие показания сервера сохраняются на диск.
	IsRestoreDefault     bool          = true                   // Значение по умолчанию для значения определяющее, загружать или нет ранее сохранённые значения из указанного файла при старте сервера.
	LogLevelDefault      string        = log.LevelErr           // Значение по умолчанию для уровня логирования.
	// CryptoKeyPathDefault string        = "/tmp/private.pem"     // Значение по умолчания для пути до файла с приватным ключом.
)

// FuncOpt Опции для конфига.
type FuncOpt func(*Config)

// StorageConfig конфигурация для хранилища.
type StorageConfig struct {
	ConnDB    string
	StorePath string
	IsRestore bool
	StoreInt  time.Duration
}

// Config Конфигурация для Агента.
type Config struct {
	Addr          string
	Key           string
	CryptoKeyPath string
	PrivateKey    *rsa.PrivateKey
	LogLevel      string
	ConfigPath    string
	StorageConfig
}

func Default() *Config {
	return &Config{
		Addr:     AddressDefault,
		LogLevel: LogLevelDefault,
		StorageConfig: StorageConfig{
			StorePath: StorePathDefault,
			StoreInt:  StoreIntervalDefault,
			IsRestore: IsRestoreDefault,
		},
		//	CryptoKeyPath: CryptoKeyPathDefault,
	}
}

// SetOpts Установка опций в Config.
func (c *Config) SetOpts(opts ...FuncOpt) {
	for i := range opts {
		opts[i](c)
	}
}

func New(opts ...FuncOpt) (*Config, error) {
	var (
		err error
	)

	cfg := Default()

	for i := range opts {
		opts[i](cfg)
	}

	// читаем приватный ключ из файла
	if cfg.CryptoKeyPath == "" {
		return cfg, nil
	}

	cfg.PrivateKey, err = crypto.RSAPrivateKey(cfg.CryptoKeyPath)
	if err != nil {
		return nil, fmt.Errorf("parse rsa key: %w", err)
	}

	return cfg, nil
}

// Установка адреса эндпоинта.
func SetAddr(addr string) FuncOpt {
	return func(cfg *Config) {
		cfg.Addr = addr
	}
}

// Установка интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск.
func SetStoreInt(interval time.Duration) FuncOpt {
	return func(cfg *Config) {
		cfg.StorageConfig.StoreInt = interval
	}
}

// Установка имени файла, куда сохраняются текущие значения.
func SetStorePath(path string) FuncOpt {
	return func(cfg *Config) {
		cfg.StorageConfig.StorePath = path
	}
}

// Установка значения, определяющее загружать или нет ранее сохранённые значения из указанного файла при старте сервера.
func SetRestore(b bool) FuncOpt {
	return func(cfg *Config) {
		cfg.StorageConfig.IsRestore = b
	}
}

// Установка строки с адресом подключения к БД.
func SetDatabaseDNS(connDB string) FuncOpt {
	return func(cfg *Config) {
		cfg.StorageConfig.ConnDB = connDB
	}
}

// Установка ключа.
func SetKey(key string) FuncOpt {
	return func(cfg *Config) {
		cfg.Key = key
	}
}

// Установка пути до файла конфигурации.
func SetConfigPath(configPath string) FuncOpt {
	return func(cfg *Config) {
		cfg.ConfigPath = configPath
	}
}

// Установка пути до файла с приватным ключом.
func SetCryptoKeyPath(cryptoKeyPath string) FuncOpt {
	return func(cfg *Config) {
		cfg.CryptoKeyPath = cryptoKeyPath
	}
}

// Установка уровня логирования.
func SetLogLevel(logLevel string) FuncOpt {
	return func(cfg *Config) {
		cfg.LogLevel = logLevel
	}
}
