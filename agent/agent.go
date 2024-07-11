// Агент для сбора рантайм-метрик и их последующей отправки на сервер по протоколу HTTP.
// Метрики собираются из пакетов runtime и gopsutil
// Полученые метрики сохраняются в хранилище [storage]
// Данные перед отправкой на сервер:
// - подписываются
// - сжимаются gzip
// - шифруются
package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/AndreyVLZ/metrics/agent/config"
	"github.com/AndreyVLZ/metrics/agent/pkg/task"
	"github.com/AndreyVLZ/metrics/agent/stats"
	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/AndreyVLZ/metrics/internal/store/inmemory"
	"github.com/AndreyVLZ/metrics/pkg/crypto"
	"github.com/AndreyVLZ/metrics/pkg/hash"
)

const (
	urlFormat             = "http://%s/updates/" // Эндпоинт для отправки метрик.
	countTask             = 3                    // Кол-во задач для Агента.
	attemptConst          = 3                    // Кол-во повторов отправки при ошибки.
	durationTaskConst     = 2                    // Таймаут опроса метрик из пакета goutils.
	retryTimeoutStepConst = 2                    // Шаг увелечения таймаута для повторной отправки.
)

// iStats Интерфейс статистики.
type iStats interface {
	Init() error
	RuntimeList() []model.Metric
	UtilList() []model.Metric
}

// storage Интерфейс хранилища.
type storage interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	AddBatch(ctx context.Context, arr []model.Metric) error
	List(ctx context.Context) ([]model.Metric, error)
}

// Агент.
type Agent struct {
	stats     iStats
	store     storage
	cfg       *config.Config
	client    *http.Client
	log       *slog.Logger
	chErr     chan error
	urlToSend string
}

// Новый Агент.
func New(cfg *config.Config, log *slog.Logger) *Agent {
	store := inmemory.New()

	return &Agent{
		cfg:       cfg,
		stats:     stats.New(),
		store:     store,
		urlToSend: fmt.Sprintf(urlFormat, cfg.Addr),
		client: &http.Client{
			Transport: &loggingRoundTripper{
				log:  log,
				next: http.DefaultTransport,
			},
		},
		log:   log,
		chErr: make(chan error),
	}
}

// Err Возвращает канал с ошибками, которые могут возникнуть при работе агента.
func (a *Agent) Err() <-chan error { return a.chErr }

// Stop Остановка агента.
func (a *Agent) Stop(ctx context.Context) error {
	arrErr := make([]error, 0, countTask+a.cfg.RateLimit+1)

	if err := a.store.Stop(ctx); err != nil {
		arrErr = append(arrErr, err)
	}

	for err := range a.Err() {
		arrErr = append(arrErr, err)
	}

	return errors.Join(arrErr...)
}

// Start Запускает агента. Возможные ошибки:
// при инициализации статистики,
// при иниицализации хранилища.
func (a *Agent) Start(ctx context.Context) error {
	a.log.DebugContext(ctx, "start agent",
		slog.String("addr", a.cfg.Addr),
		slog.Group("flags",
			slog.String("confgigPath", a.cfg.ConfigPath),
			slog.String("publicKeyPath", a.cfg.CryptoKeyPath),
			slog.String("poolInt", a.cfg.PollInterval.String()),
			slog.String("reportInt", a.cfg.ReportInterval.String()),
			slog.Int("rateLimit", a.cfg.RateLimit),
			slog.String("key", string(a.cfg.Key)),
			slog.String("lvl", a.cfg.LogLevel),
		),
	)

	if err := a.stats.Init(); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := a.store.Start(ctx); err != nil {
		return fmt.Errorf("%w", err)
	}

	go a.start(ctx)

	return nil
}

// start Запуск task'ов и worker'ов Агента.
func (a *Agent) start(ctx context.Context) {
	ctxCan, cancel := context.WithCancel(ctx)
	defer cancel()

	fnSend := func(arr []model.Metric) error { return a.sendBatch(ctxCan, arr) }
	// запуск задач для Агента
	chList := a.runTaskPoll(ctxCan)
	// запускаем воркеры
	chErrWorker := runWorkerPool(a.cfg.RateLimit, chList, fnSend, a.log)

	// ждем закрытие канала с ошибками
	// закрытый канал сигнализирует о том, что все worker'ы и task'и остановились
	// и можно закрывать канал a.chErr
	for err := range chErrWorker {
		cancel() // при ошибки отменяем контекст taskPoll
		a.chErr <- err
	}

	close(a.chErr)
}

// sendBatch Отправка метрик.
func (a *Agent) sendBatch(ctx context.Context, arr []model.Metric) error {
	var header http.Header = make(map[string][]string)

	data, err := json.Marshal(model.BuildArrMetricJSON(arr))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// хeшируем данные
	sum, err := hashed(a.cfg.Key, data)
	if err != nil {
		return fmt.Errorf("req hashed: %w", err)
	}

	header.Set("HashSHA256", hex.EncodeToString(sum))

	// сжимаем данные
	dataCompress, err := gzipCompres(data)
	if err != nil {
		return fmt.Errorf("req compress: %w", err)
	}

	header.Set("Content-Encoding", "gzip")

	// шифруем данные
	dataEncrypt, err := encrypt(a.cfg.PublicKey, dataCompress)
	if err != nil {
		return fmt.Errorf("req crypto: %w", err)
	}

	return retry(ctx, attemptConst, time.Second, a.log, func() error {
		// собираем запрос
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.urlToSend, bytes.NewReader(dataEncrypt))
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}

		header.Set("Content-Type", "application/json")
		clientIP, err := getIP()
		if err != nil {
			return err
		}
		header.Set("X-Real-IP", clientIP.String())

		req.Header = header

		return a.do(req)
	})
}

// Повторный вызов функции fnSend attempts раз.
// Возвращает ошибку с кол-вом вызовов функции и самой ошибкой fnSend.
func retry(ctx context.Context, attempts int, sleep time.Duration, log *slog.Logger, fnSend func() error) error {
	var err error

	for i := 0; i <= attempts; i++ {
		if i > 0 {
			log.DebugContext(ctx, "send",
				slog.Group("send",
					slog.Int("retry", i),
					slog.String("error", err.Error()),
				),
			)

			select {
			case <-ctx.Done():
				return nil
			case <-time.After(sleep):
				sleep += retryTimeoutStepConst * time.Second
			}
		}

		err = fnSend()
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("попыток %d, error: %w", attempts, err)
}

// hashed Возвращает хеш.
func hashed(key, data []byte) ([]byte, error) {
	if len(key) == 0 {
		return data, nil
	}

	sum, err := hash.SHA256(data, key)
	if err != nil {
		return nil, fmt.Errorf("hash: %w", err)
	}

	return sum, nil
}

// gzipCompres Сжимает данные.
func gzipCompres(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	if _, err := gzipWriter.Write(data); err != nil {
		return nil, fmt.Errorf("gzip write: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("gzip close: %w", err)
	}

	return buf.Bytes(), nil
}

// encrypt Шифрует данные публичным ключом.
func encrypt(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	if publicKey == nil {
		return data, nil
	}

	cipher, err := crypto.Encrypt(publicKey, data)
	if err != nil {
		return nil, fmt.Errorf("encrypt len[%d] :%w", len(data), err)
	}

	return cipher, nil
}

// do Выполняет запрос.
func (a *Agent) do(req *http.Request) error {
	// выполняем запрос
	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("err to send request: %w", err)
	}

	// закрывает тело ответа
	if err = resp.Body.Close(); err != nil {
		return fmt.Errorf("err body close: %w", err)
	}

	return nil
}

// runTaskPoll Запуск пула задач Агента.
func (a *Agent) runTaskPoll(ctx context.Context) <-chan []model.Metric {
	ctxCan, cancel := context.WithCancel(ctx)
	chList := make(chan []model.Metric)

	taskPoll := task.NewPoll(countTask, a.log)
	// определяем задачи для Агента
	taskPoll.Add(
		task.New("update runtime", // сбор метрик (опрос runtime)
			a.cfg.PollInterval,
			func() error { return a.store.AddBatch(ctxCan, a.stats.RuntimeList()) },
		),
		task.New("update gopsutil", // сбор метрики из пакета gopsutil
			a.cfg.ReportInterval/durationTaskConst,
			func() error { return a.store.AddBatch(ctxCan, a.stats.UtilList()) },
		),
		task.New("read from store", // чтение метрик из store
			a.cfg.ReportInterval,
			func() error {
				list, err := a.store.List(ctxCan)
				if err != nil {
					return fmt.Errorf("%w", err)
				}
				chList <- list

				return nil
			},
		),
	)

	// запускаем пул задач
	chErrTask := taskPoll.Run(ctxCan)

	// в отдельной горутине ждем закрытие канала с ошибками
	// закрытый канал сигнализирует о том, что все task'и остановились
	// и можно закрывать канал chList
	go func() {
		for err := range chErrTask {
			cancel() // при ошибке отменяем контекст для taskPoll
			a.chErr <- err
		}
		close(chList)
	}()

	return chList
}

// runWorkerPool Запуск пула worker'ов.
func runWorkerPool(rateLimit int, jobc <-chan []model.Metric, fnSend func([]model.Metric) error, log *slog.Logger) <-chan error {
	var wgPool sync.WaitGroup

	errc := make(chan error)

	for idWorker := 0; idWorker < rateLimit; idWorker++ {
		wgPool.Add(1)

		go func() {
			for list := range jobc {
				if err := fnSend(list); err != nil {
					errc <- err
				}
			}

			wgPool.Done()
		}()
	}

	// в отдельной горутине ждем остановки worker'ов
	go func() {
		wgPool.Wait()
		close(errc) // закрываем канал, в который писали worker'ы
		log.Debug("worker poll", "wait", "ok")
	}()

	return errc
}

func getIP() (net.IP, error) {
	addrs, err := net.LookupHost("localhost")
	if err != nil {
		return nil, fmt.Errorf("lookupHost: %w", err)
	}

	if len(addrs) == 0 {
		return nil, errors.New("no find addrs")
	}

	netIP := net.ParseIP(addrs[0])
	if netIP == nil {
		return nil, fmt.Errorf("parse IP [%s]: %w", addrs[0], err)
	}

	return netIP, nil
}
