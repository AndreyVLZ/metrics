package agent

/*
import (
	"flag"

	"github.com/AndreyVLZ/metrics/agent/config"
	"github.com/AndreyVLZ/metrics/pkg/env"
)

func ParseFlag() []config.FuncOpt {
	addrPtr := flag.String("a", AddressDefault, "адрес эндпоинта HTTP-сервера")
	pollIntervarPtr := flag.Int("p", PollIntervalDefault, "частота опроса метрик из пакета runtime")
	reportIntervalPtr := flag.Int("r", ReportIntervalDefault, "частота отправки метрик на сервер")
	keyPtr := flag.String("k", KeyDafault, "ключ")
	rateLimitPtr := flag.Int("l", RateLimitDefault, "количество одновременно исходящих запросов на сервер")
	cryproKeyPtr := flag.String("crypto-key", CryproKeyPathDefault, "путь до файла с публичным ключом")

	return []config.FuncOpt{
		config.SetAddr(*addrPtr),
		config.SetPollInterval(*pollIntervarPtr),
		config.SetReportInterval(*reportIntervalPtr),
		config.SetKey(*keyPtr),
		config.SetRateLimit(*rateLimitPtr),
		config.SetCryptoKeyPath(*cryproKeyPtr),
	}
}

func ParseEnv() []config.FuncOpt {
	var (
		addr           string
		pollInterval   int
		reportInterval int
		key            string
		rateLimit      int
		cryptoKeyPath  string
	)

	opts := make([]config.FuncOpt, 0)
	if err := env.String(&addr, "ADDRESS")(); err == nil {
		opts = append(opts, config.SetAddr(addr))
	}

	if err := env.Int(&pollInterval, "POLL_INTERVAL")(); err == nil {
		opts = append(opts, config.SetPollInterval(pollInterval))
	}

	if err := env.Int(&reportInterval, "REPORT_INTERVAL")(); err == nil {
		opts = append(opts, config.SetReportInterval(reportInterval))
	}

	if err := env.String(&key, "KEY")(); err == nil {
		opts = append(opts, config.SetKey(key))
	}

	if err := env.Int(&rateLimit, "RATE_LIMIT")(); err == nil {
		opts = append(opts, config.SetRateLimit(rateLimit))
	}

	if err := env.String(&cryptoKeyPath, "CRYPTO_KEY")(); err == nil {
		opts = append(opts, config.SetCryptoKeyPath(cryptoKeyPath))
	}

	return opts
}

func Parse() ([]config.FuncOpt, error) {
	opts := make([]config.FuncOpt, 0)
	addr := flagenv.String("ADDRESS", "a", AddressDefault, "адрес эндпоинта HTTP-сервера")
	opts = append(opts, config.SetAddr(*addr))

	return opts, nil
}
*/
