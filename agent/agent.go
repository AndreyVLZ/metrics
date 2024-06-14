package agent

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/AndreyVLZ/metrics/agent/stats"
	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/AndreyVLZ/metrics/internal/store/inmemory"
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
	cfg       *config
	client    *http.Client
	log       *slog.Logger
	chErr     chan error
	urlToSend string
}

// Новый Агент.
func New(log *slog.Logger, fnOpts ...FuncOpt) *Agent {
	cfg := newConfig(fnOpts...)
	store := inmemory.New()

	return &Agent{
		cfg:       cfg,
		stats:     stats.New(),
		store:     store,
		urlToSend: fmt.Sprintf(urlFormat, cfg.addr),
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

// Отправка данных.
func (a *Agent) sendData(ctx context.Context, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.urlToSend, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("err build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if len(data) != 0 {
		sum, errHash := hash.SHA256(data, []byte(a.cfg.key))
		if errHash != nil {
			return fmt.Errorf("hash: %w", errHash)
		}

		req.Header.Set("HashSHA256", hex.EncodeToString(sum))
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("err to send request: %w", err)
	}

	if err = resp.Body.Close(); err != nil {
		return fmt.Errorf("err body close: %w", err)
	}

	return nil
}

// Отправка метрик.
func (a *Agent) send(ctx context.Context, arr []model.Metric) error {
	data, err := json.Marshal(model.BuildArrMetricJSON(arr))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return retry(ctx, attemptConst, time.Second, a.log, func() error {
		return a.sendData(ctx, data)
	})
}

// Возвращает канал с ошибками, которые могут возникнуть при работе агента.
func (a *Agent) Err() <-chan error { return a.chErr }

// Остановка агента.
func (a *Agent) Stop(ctx context.Context) error {
	arrErr := make([]error, 0, countTask+a.cfg.rateLimit+1)

	if err := a.store.Stop(ctx); err != nil {
		arrErr = append(arrErr, err)
	}

	for err := range a.Err() {
		arrErr = append(arrErr, err)
	}

	return errors.Join(arrErr...)
}

// Старт агента. Возможные ошибки:
// при инициализации статистики,
// при иниицализации хранилища.
func (a *Agent) Start(ctx context.Context) error {
	a.log.DebugContext(ctx, "start agent",
		slog.String("addr", a.cfg.addr),
		slog.Group("flags",
			slog.Int("poolInt", a.cfg.pollInterval),
			slog.Int("reportInt", a.cfg.reportInterval),
			slog.Int("rateLimit", a.cfg.rateLimit),
			slog.String("key", a.cfg.key),
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

// Запуск task'ов и worker'ов Агента.
func (a *Agent) start(ctx context.Context) {
	ctxCan, cancel := context.WithCancel(ctx)
	defer cancel()

	fnSend := func(arr []model.Metric) error { return a.send(ctxCan, arr) }
	chList, chErrTask := a.runTaskPool(ctxCan)
	chErrWorker := runWorkerPool(a.cfg.rateLimit, chList, fnSend)

	chErr := fanIn(chErrWorker, chErrTask)

	for err := range chErr {
		cancel() // При возникновении ошибки отменяем TaskPool
		a.chErr <- err
	}

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
			duration: time.Duration(a.cfg.pollInterval) * time.Second,
			fn:       func() error { return a.store.AddBatch(ctx, a.stats.RuntimeList()) },
		},
		{
			name:     "update gopsutil", // сбор метрики из пакета gopsutil
			duration: time.Duration(a.cfg.reportInterval/durationTaskConst) * time.Second,
			fn:       func() error { return a.store.AddBatch(ctx, a.stats.UtilList()) },
		},
		{
			name:     "read from store", // чтение метрик из store
			duration: time.Duration(a.cfg.reportInterval) * time.Second,
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

// Запуск worker'ов.
func runWorkerPool(rateLimit int, jobc <-chan []model.Metric, fnSend func([]model.Metric) error) <-chan error {
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

	go func() {
		wgPool.Wait()
		close(errc)
	}()

	return errc
}
