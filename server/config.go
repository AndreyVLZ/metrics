package server

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/AndreyVLZ/metrics/pkg/crypto"
)

const (
	AddressDefault       string = "localhost:8080"       // Значение по умолчанию для адреса эндпоинта HTTP-сервера.
	StoreIntervalDefault int    = 300                    // Значение по умолчанию для интервала времени в секундах, по истечении которого текущие показания сервера сохраняются на диск.
	StoragePathDefault   string = "/tmp/metrics-db.json" // Значение по умолчанию для имени файла, куда сохраняются текущие значения.
	IsRestoreDefault     bool   = true                   // Значение по умолчанию для значения определяющее, загружать или нет ранее сохранённые значения из указанного файла при старте сервера.
	CryptoKeyPathDefault string = "/tmp/private.pem"     // Значение по умолчания для пути до файла с приватным ключом.
)

// Конфиг для Сервера.
type Config struct {
	addr          string          `json:"address"`
	storePath     string          `json:"store_file"`
	dbDNS         string          `json:"database_dsn"`
	key           string          `json:"key"`
	storeInt      time.Duration   `json:"store_interval"`
	isRestore     bool            `json:"restore"`
	cryptoKeyPath string          `json:"crypto_key"`
	privateKey    *rsa.PrivateKey `json:"-"`
	configPath    string          `json:"-"`
}

// Опции кофига.
type FuncOpt func(*Config)

// Новый конфиг с установленными опциями.
func NewConfig(configPath string, opts ...FuncOpt) (*Config, error) {
	var err error

	// конфиг по умолчанию
	cfg := Config{
		addr:          AddressDefault,
		storeInt:      time.Duration(StoreIntervalDefault) * time.Second,
		storePath:     StoragePathDefault,
		isRestore:     IsRestoreDefault,
		dbDNS:         StoragePathDefault,
		cryptoKeyPath: CryptoKeyPathDefault,
	}

	// конфиг из файла
	if configPath != "" {
		cfg, err = configFromFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("config from file: %w", err)
		}
	}

	// устанавливаем переданные опции в конфиг
	for i := range opts {
		opts[i](&cfg)
	}

	if cfg.cryptoKeyPath == "" {
		return &cfg, nil
	}

	// читаем приватный ключ из файла
	cfg.privateKey, err = crypto.RSAPrivateKey(cfg.cryptoKeyPath)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	return &cfg, nil
}

func configFromFile(filePath string) (Config, error) {
	var cfg Config

	fileByte, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("open config file:%w", err)
	}

	if err := json.Unmarshal(fileByte, &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal: %w", err)
	}

	return cfg, nil
}

/*
func NewConfig(opts ...FuncOpt) *config {
	cfg := config{
		addr:          AddressDefault,
		storeInt:      time.Duration(StoreIntervalDefault) * time.Second,
		storePath:     StoragePathDefault,
		isRestore:     IsRestoreDefault,
		dbDNS:         StoragePathDefault,
		cryptoKeyPath: CryptoKeyPathDefault,
	}

	for i := range opts {
		opts[i](&cfg)
	}

	return &cfg
}
// Установка пути до файла конфигурации.
func SetConfigFilePath(configPath string) FuncOpt {
	return func(cfg *Config) {
		cfg.configPath = configPath
	}
}
*/

// Установка адреса эндпоинта.
func SetAddr(addr string) FuncOpt {
	return func(cfg *Config) {
		cfg.addr = addr
	}
}

// Установка интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск.
func SetStoreInt(interval int) FuncOpt {
	return func(cfg *Config) {
		cfg.storeInt = time.Duration(interval) * time.Second
	}
}

// Установка имени файла, куда сохраняются текущие значения.
func SetStorePath(path string) FuncOpt {
	return func(cfg *Config) {
		cfg.storePath = path
	}
}

// Установка значения, определяющее загружать или нет ранее сохранённые значения из указанного файла при старте сервера.
func SetRestore(b bool) FuncOpt {
	return func(cfg *Config) {
		cfg.isRestore = b
	}
}

// Установка строки с адресом подключения к БД.
func SetDatabaseDNS(dns string) FuncOpt {
	return func(cfg *Config) {
		cfg.dbDNS = dns
	}
}

// Установка ключа.
func SetKey(key string) FuncOpt {
	return func(cfg *Config) {
		cfg.key = key
	}
}

// Установка пути до файла с приватным ключом.
func SetCryptoKeyPath(cryptoKeyPath string) FuncOpt {
	return func(cfg *Config) {
		cfg.cryptoKeyPath = cryptoKeyPath
	}
}
