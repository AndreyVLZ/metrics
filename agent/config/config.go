package config

import (
	"crypto/rsa"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
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
)

// Опции для конфига.
type FuncOpt func(cfg *Config)

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
	fs.Func("p", "частота опроса метрик из пакета runtime", func(flagPollStr string) error {
		poll, err := time.ParseDuration(flagPollStr)
		if err != nil {
			return fmt.Errorf("parse poll interval: %w", err)
		}

		flagOpts = append(flagOpts, SetPollInterval(poll))

		return nil
	})
	fs.Func("r", "частота отправки метрик на сервер", func(flagReportStr string) error {
		report, err := time.ParseDuration(flagReportStr)
		if err != nil {
			return fmt.Errorf("parse report interval: %w", err)
		}

		flagOpts = append(flagOpts, SetReportInterval(report))

		return nil
	})
	fs.Func("l", "количество одновременно исходящих запросов на сервер", func(flagRateStr string) error {
		rate, err := strconv.Atoi(flagRateStr)
		if err != nil {
			return fmt.Errorf("parse rate limit: %w", err)
		}

		flagOpts = append(flagOpts, SetRateLimit(rate))

		return nil
	})
	fs.Func("a", "адрес эндпоинта HTTP-сервера", func(flagAddress string) error {
		flagOpts = append(flagOpts, SetAddr(flagAddress))

		return nil
	})
	fs.Func("k", "ключ", func(flagKey string) error {
		flagOpts = append(flagOpts, SetKey(flagKey))

		return nil
	})
	fs.Func("crypto-key", "путь до файла с приватным ключом", func(flagCryptoKeyPath string) error {
		flagOpts = append(flagOpts, SetCryptoKeyPath(flagCryptoKeyPath))

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

	rateStr, isExist := os.LookupEnv("RATE_LIMIT")
	if isExist {
		rateENV, err := strconv.Atoi(rateStr)
		if err != nil {
			return nil, fmt.Errorf("parse env RATE_LIMIT: %w", err)
		}

		opts = append(opts, SetRateLimit(rateENV))
	}

	pollStr, isExist := os.LookupEnv("POLL_INTERVAL")
	if isExist {
		pollENV, err := strconv.Atoi(pollStr)
		if err != nil {
			return nil, fmt.Errorf("parse env POLL_INTERVAL: %w", err)
		}

		opts = append(opts, SetPollInterval(time.Duration(pollENV)*time.Second))
	}

	reportStr, isExist := os.LookupEnv("REPORT_INTERVAL")
	if isExist {
		reportENV, err := strconv.Atoi(reportStr)
		if err != nil {
			return nil, fmt.Errorf("parse env REPORT_INTERVAL: %w", err)
		}

		opts = append(opts, SetReportInterval(time.Duration(reportENV)*time.Second))
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
		Addr           *string `json:"address"`         // аналог переменной окружения ADDRESS или флага -a
		ReportInterval *string `json:"report_interval"` // аналог переменной окружения REPORT_INTERVAL или флага -r
		PollInterval   *string `json:"poll_interval"`   // аналог переменной окружения POLL_INTERVAL или флага -p
		CryptoKeyPath  *string `json:"crypto_key"`      // аналог переменной окружения CRYPTO_KEY или флага -crypto-key
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

	if inConfig.CryptoKeyPath != nil {
		opts = append(opts, SetCryptoKeyPath(*inConfig.CryptoKeyPath))
	}

	if inConfig.ReportInterval != nil {
		report, err := time.ParseDuration(*inConfig.ReportInterval)
		if err != nil {
			return nil, fmt.Errorf("parse duration: %w", err)
		}

		opts = append(opts, SetReportInterval(report))
	}

	if inConfig.PollInterval != nil {
		poll, err := time.ParseDuration(*inConfig.PollInterval)
		if err != nil {
			return nil, fmt.Errorf("parse duration: %w", err)
		}

		opts = append(opts, SetPollInterval(poll))
	}

	return opts, nil
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
