package config

import (
	"crypto/rsa"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
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
	TrustedSubnet net.IP
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
	}
}

// New ...
func New(args []string) (*Config, error) {
	cfg := Default()

	flagOpts, err := cfg.fromFlag(args)
	if err != nil {
		return nil, err
	}

	envOpts, err := cfg.fromENV()
	if err != nil {
		return nil, err
	}

	fileOpts, err := optsFromFile(cfg.ConfigPath)
	if err != nil {
		return nil, err
	}

	cfg.SetOpts(fileOpts...)
	cfg.SetOpts(flagOpts...)
	cfg.SetOpts(envOpts...)

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

// SetOpts Установка опций в Config.
func (c *Config) SetOpts(opts ...FuncOpt) {
	for i := range opts {
		opts[i](c)
	}
}

// fromFlag Опции для конфига из флагов.
func (c *Config) fromFlag(args []string) ([]FuncOpt, error) {
	flagOpts := make([]FuncOpt, 0)

	fs := flag.NewFlagSet(args[0], flag.ExitOnError)

	fs.StringVar(&c.ConfigPath, "c", "", "путь до файла конфигурации")
	fs.Func("i", "интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск", func(flagStoreIntervalStr string) error {
		storeInterval, err := time.ParseDuration(flagStoreIntervalStr)
		if err != nil {
			return fmt.Errorf("parse poll interval: %w", err)
		}

		flagOpts = append(flagOpts, SetStoreInt(storeInterval))

		return nil
	})
	fs.Func("a", "адрес эндпоинта HTTP-сервера", func(flagAddress string) error {
		flagOpts = append(flagOpts, SetAddr(flagAddress))

		return nil
	})
	fs.Func("t", "subnet", func(flagTrustedSubnetStr string) error {
		netIP := net.ParseIP(flagTrustedSubnetStr)
		flagOpts = append(flagOpts, SetTrustedSubnet(netIP))

		return nil
	})
	fs.Func("f", "полное имя файла, куда сохраняются текущие значения", func(flagFileStorePath string) error {
		flagOpts = append(flagOpts, SetStorePath(flagFileStorePath))

		return nil
	})
	fs.Func("d", "строка с адресом подключения к БД", func(flagConnDB string) error {
		flagOpts = append(flagOpts, SetDatabaseDSN(flagConnDB))

		return nil
	})
	fs.Func("r", "определяющее, загружать или нет ранее сохранённые значения из указанного файла при старте сервера", func(flagIsRestoreStr string) error {
		isRestore, err := strconv.ParseBool(flagIsRestoreStr)
		if err != nil {
			return fmt.Errorf("flag restore: %w", err)
		}

		flagOpts = append(flagOpts, SetRestore(isRestore))

		return nil
	})
	fs.Func("crypto-key", "путь до файла с приватным ключом", func(flagCryptoKeyPath string) error {
		flagOpts = append(flagOpts, SetCryptoKeyPath(flagCryptoKeyPath))

		return nil
	})
	fs.Func("k", "ключ", func(flagKey string) error {
		flagOpts = append(flagOpts, SetKey(flagKey))

		return nil
	})
	fs.Func("lvl", "уровень логирования", func(flagLogLevel string) error {
		flagOpts = append(flagOpts, SetLogLevel(flagLogLevel))

		return nil
	})

	if err := fs.Parse(args[1:]); err != nil {
		return nil, fmt.Errorf("parse flags: %w", err)
	}

	return flagOpts, nil
}

// fromENV Опции для конфига из ENV.
func (c *Config) fromENV() ([]FuncOpt, error) {
	opts := make([]FuncOpt, 0)

	configPathENV, isExist := os.LookupEnv("CONFIG")
	if isExist {
		c.ConfigPath = configPathENV
	}

	trustedSubnet, isExist := os.LookupEnv("TRUSTED_SUBNET")
	if isExist {
		netIP := net.ParseIP(trustedSubnet)
		opts = append(opts, SetTrustedSubnet(netIP))
	}

	addrENV, isExist := os.LookupEnv("ADDRESS")
	if isExist {
		opts = append(opts, SetAddr(addrENV))
	}

	keyENV, isExist := os.LookupEnv("KEY")
	if isExist {
		opts = append(opts, SetKey(keyENV))
	}

	cryptoKeyPathENV, isExist := os.LookupEnv("CRYPTO_KEY")
	if isExist {
		opts = append(opts, SetCryptoKeyPath(cryptoKeyPathENV))
	}

	logLevelENV, isExist := os.LookupEnv("LVL")
	if isExist {
		opts = append(opts, SetLogLevel(logLevelENV))
	}

	storeIntStr, isExist := os.LookupEnv("STORE_INTERVAL")
	if isExist {
		storeIntENV, err := time.ParseDuration(storeIntStr)
		if err != nil {
			return nil, fmt.Errorf("parse env STORE_INTERVAL: %w", err)
		}

		opts = append(opts, SetStoreInt(storeIntENV))
	}

	fileStoragePathENV, isExist := os.LookupEnv("FILE_STORAGE_PATH")
	if isExist {
		opts = append(opts, SetStorePath(fileStoragePathENV))
	}

	dsnENV, isExist := os.LookupEnv("DATABASE_DSN")
	if isExist {
		opts = append(opts, SetDatabaseDSN(dsnENV))
	}

	isRestoreStr, isExist := os.LookupEnv("RESTORE")
	if isExist {
		isRestoreENV, err := strconv.ParseBool(isRestoreStr)
		if err != nil {
			return nil, fmt.Errorf("parse env RESTORE: %w", err)
		}

		opts = append(opts, SetRestore(isRestoreENV))
	}

	return opts, nil
}

// optsFromFile Опции для конфига из файла.
func optsFromFile(configPath string) ([]FuncOpt, error) {
	if configPath == "" {
		return []FuncOpt{}, nil
	}

	opts := make([]FuncOpt, 0)

	inConfig := struct {
		Addr          *string `json:"address"`        // аналог переменной окружения ADDRESS или флага -a
		IsRestore     *bool   `json:"restore"`        // аналог переменной окружения RESTORE или флага -r
		StoreInterval *string `json:"store_interval"` // аналог переменной окружения STORE_INTERVAL или флага -i
		FileStorePath *string `json:"store_file"`     // аналог переменной окружения STORE_FILE или -f
		ConnDB        *string `json:"database_dsn"`   // аналог переменной окружения DATABASE_DSN или флага -d
		CryptoKeyPath *string `json:"crypto_key"`     // аналог переменной окружения CRYPTO_KEY или флага -crypto-key
		TrustedSubnet *string `json:"trusted_subnet"` // аналог переменной окружения TRUSTED_SUBNET или флага -t
	}{}

	fileByte, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	if err := json.Unmarshal(fileByte, &inConfig); err != nil {
		return nil, fmt.Errorf("unmarshal :%w", err)
	}

	// Добавляем опции если только они были явно установленный в файле
	if inConfig.Addr != nil {
		opts = append(opts, SetAddr(*inConfig.Addr))
	}

	if inConfig.IsRestore != nil {
		opts = append(opts, SetRestore(*inConfig.IsRestore))
	}

	if inConfig.TrustedSubnet != nil {
		netIP := net.ParseIP(*inConfig.TrustedSubnet)
		opts = append(opts, SetTrustedSubnet(netIP))
	}

	if inConfig.StoreInterval != nil {
		storeInt, err := time.ParseDuration(*inConfig.StoreInterval)
		if err != nil {
			return nil, fmt.Errorf("parse duration: %w", err)
		}

		opts = append(opts, SetStoreInt(storeInt))
	}

	if inConfig.FileStorePath != nil {
		opts = append(opts, SetStorePath(*inConfig.FileStorePath))
	}

	if inConfig.ConnDB != nil {
		opts = append(opts, SetDatabaseDSN(*inConfig.ConnDB))
	}

	if inConfig.CryptoKeyPath != nil {
		opts = append(opts, SetCryptoKeyPath(*inConfig.CryptoKeyPath))
	}

	return opts, nil
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
func SetDatabaseDSN(connDB string) FuncOpt {
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

// Установка уровня логирования.
func SetTrustedSubnet(trustedSubnet net.IP) FuncOpt {
	return func(cfg *Config) {
		cfg.TrustedSubnet = trustedSubnet
	}
}
