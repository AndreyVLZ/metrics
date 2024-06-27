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

// Интерфейс статистики.
type iStats interface {
	Init() error
	RuntimeList() []model.Metric
	UtilList() []model.Metric
}

// Интерфейс хранилища.
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
func New(log *slog.Logger, cfg *config.Config) *Agent {
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

// hashed
func hashed(key, data []byte) ([]byte, error) {
	if len(key) == 0 {
		// return nil, errors.New("key empty")
	}

	sum, err := hash.SHA256(data, key)
	if err != nil {
		return nil, fmt.Errorf("hash: %w", err)
	}

	return sum, nil
}

// gzipCompres сжимает данные.
func gzipCompres(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	if _, err := gzipWriter.Write(data); err != nil {
		return nil, fmt.Errorf("gzip write: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("gzip close: %w", err)
	}

	fmt.Printf("compressed [%d] to [%d] bytes\r\n", len(data), buf.Len())

	return buf.Bytes(), nil
}

func encrypt(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	if publicKey == nil {
		return nil, errors.New("public key empty")
	}

	cipher, err := crypto.Encrypt(publicKey, data)
	if err != nil {
		return nil, fmt.Errorf("encrypt len[%d] :%w", len(data), err)
	}

	return cipher, nil
}

// Отправка метрик.
func (a *Agent) sendBatch(ctx context.Context, arr []model.Metric) error {
	data, err := json.Marshal(model.BuildArrMetricJSON(arr))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	req, err := a.buildRequest(ctx, data)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return retry(ctx, attemptConst, time.Second, a.log, func() error {
		//return a.sendData(ctx, data)
		return a.do(req)
	})
}

func (a *Agent) buildRequest(ctx context.Context, data []byte) (*http.Request, error) {
	var header http.Header = make(map[string][]string)

	// хeшируем данные
	sum, err := hashed(a.cfg.Key, data)
	if err != nil {
		return nil, fmt.Errorf("req hashed: %w", err)
	}

	header.Set("HashSHA256", hex.EncodeToString(sum))

	// сжимаем данные
	dataCompress, err := gzipCompres(data)
	if err != nil {
		return nil, fmt.Errorf("req compress: %w", err)
	}

	header.Set("Content-Encoding", "gzip")

	//dataEncrypt := dataCompress
	// шифруем данные
	dataEncrypt, err := encrypt(a.cfg.PublicKey, dataCompress)
	if err != nil {
		return nil, fmt.Errorf("req crypto: %w", err)
	}

	// собираем запрос
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.urlToSend, bytes.NewReader(dataEncrypt))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	header.Set("Content-Type", "application/json")
	req.Header = header

	return req, nil
}

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
			slog.Int("poolInt", a.cfg.PollInterval),
			slog.Int("reportInt", a.cfg.ReportInterval),
			slog.Int("rateLimit", a.cfg.RateLimit),
			slog.String("key", string(a.cfg.Key)),
			slog.String("publicKeyPath", a.cfg.CryptoKeyPath),
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

// runTaskPoll Запуск пула задач Агента.
func (a *Agent) runTaskPoll(ctx context.Context) <-chan []model.Metric {
	ctxCan, cancel := context.WithCancel(ctx)
	chList := make(chan []model.Metric)

	taskPoll := task.NewPoll(3, a.log)
	// определяем задачи для Агента
	taskPoll.Add(
		task.New("update runtime", // сбор метрик (опрос runtime)
			time.Duration(a.cfg.PollInterval)*time.Second,
			func() error { return a.store.AddBatch(ctxCan, a.stats.RuntimeList()) },
		),
		task.New("update gopsutil", // сбор метрики из пакета gopsutil
			time.Duration(a.cfg.ReportInterval/durationTaskConst)*time.Second,
			func() error { return a.store.AddBatch(ctxCan, a.stats.UtilList()) },
		),
		task.New("read from store", // чтение метрик из store
			time.Duration(a.cfg.ReportInterval)*time.Second,
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

/*
func (a *Agent) start1(ctx context.Context) {
	//chList, chErrTask := a.runTaskPool(ctxCan)
	ctxCan, cancel := context.WithCancel(ctx)
	defer cancel()

	chList := make(chan []model.Metric)

	fnSend := func(arr []model.Metric) error { return a.sendBatch(ctxCan, arr) }

	taskPoll := tasks.NewPoll(3, a.log)
	taskPoll.Add(
		tasks.New("update runtime",
			time.Duration(a.cfg.PollInterval)*time.Second,
			func() error { return a.store.AddBatch(ctxCan, a.stats.RuntimeList()) },
		),
		tasks.New("update gopsutil",
			time.Duration(a.cfg.ReportInterval/durationTaskConst)*time.Second,
			func() error { return a.store.AddBatch(ctxCan, a.stats.UtilList()) },
		),
		tasks.New("read from store",
			time.Duration(a.cfg.ReportInterval)*time.Second,
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

	// запускаем задачи
	chErrTask := taskPoll.Run(ctxCan)
	// запускаем воркеры
	chErrWorker := runWorkerPool(a.cfg.RateLimit, chList, fnSend)
	// объединяем два канала с ошибками в один
	chErr := fanIn(chErrWorker, chErrTask)

	for err := range chErr {
		// taskPoll ждет отмены контекста ctxCan
		// после остановки всех task'ов закрывает канал chErrTask
		cancel() // При возникновении ошибки отменяем TaskPool
		a.chErr <- err
	}

	// workerPoll ждет когда закроется канал chList,
	close(chList)
	close(a.chErr)
}
// Задача для Агента.
type task struct {
	fn       func() error
	name     string
	duration time.Duration
}

// Запуск пула задач для агента.
func (a *Agent) runTaskPool(ctx context.Context) (<-chan []model.Metric, <-chan error) {
	chErr := make(chan error)
	chList := make(chan []model.Metric)

	tasks := []task{
		{
			name:     "update runtime", // сбор метрик (опрос runtime)
			duration: time.Duration(a.cfg.PollInterval) * time.Second,
			fn:       func() error { return a.store.AddBatch(ctx, a.stats.RuntimeList()) },
		},
		{
			name:     "update gopsutil", // сбор метрики из пакета gopsutil
			duration: time.Duration(a.cfg.ReportInterval/durationTaskConst) * time.Second,
			fn:       func() error { return a.store.AddBatch(ctx, a.stats.UtilList()) },
		},
		{
			name:     "read from store", // чтение метрик из store
			duration: time.Duration(a.cfg.ReportInterval) * time.Second,
			fn: func() error {
				list, err := a.store.List(ctx)
				if err != nil {
					return fmt.Errorf("%w", err)
				}

				chList <- list

				return nil
			},
		},
	}

	var wgTask sync.WaitGroup

	for idTask := range tasks {
		wgTask.Add(1)

		go func(task task) {
			a.log.Debug("запуск задачи", "name", task.name)
			defer wgTask.Done()

			for {
				select {
				case <-ctx.Done():
					a.log.Debug("отмена задачи", "name", task.name)
					return
				case <-time.After(task.duration):
					if err := task.fn(); err != nil {
						chErr <- err
					}
				}
			}
		}(tasks[idTask])
	}

	go func() {
		wgTask.Wait() // Ждем завершения работы всех задач [Task]
		close(chList) // Закрываем канал в который писали Task'и
		close(chErr)  // Закрываем канал в который писали Task'и
	}()

	return chList, chErr
}

// fanIn объединяет несколько каналов в один.
func fanIn(errChs ...<-chan error) <-chan error {
	var wg sync.WaitGroup

	finalCh := make(chan error)

	for _, chErr := range errChs {
		wg.Add(1)

		go func(chClosure <-chan error) {
			defer wg.Done()

			for data := range chClosure {
				finalCh <- data
			}
		}(chErr)
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}
*/
