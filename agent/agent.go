package agent

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/AndreyVLZ/metrics/agent/stats"
	"github.com/AndreyVLZ/metrics/internal/hash"
	"github.com/AndreyVLZ/metrics/internal/model"
)

const (
	urlFormat = "http://%s/updates/" // Эндпоинт для отправки метрик.
	countTask = 3                    // Кол-во задач для Агента.
)

const (
	attemptConst          = 3 // Кол-во повторов отправки при ошибки.
	durationTaskConst     = 2 // Таймаут опроса метрик из пакета goutils.
	retryTimeoutStepConst = 2 // Шаг увелечения таймаута для повторной отправки.
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
	cfg       *Config
	stats     iStats
	store     storage
	urlToSend string
	client    *http.Client
	chErr     chan error
}

// Новый Агент.
func New(cfg *Config, store storage, log *slog.Logger) *Agent {
	return &Agent{
		cfg:       cfg,
		stats:     stats.New(),
		store:     store,
		urlToSend: fmt.Sprintf(urlFormat, cfg.addr),
		client:    http.DefaultClient,
		/*
			client: &http.Client{
				Transport: &loggingRoundTripper{
					log: log,
					//next: &retryRoundTripper{
					next: http.DefaultTransport,
					//	maxRetries:     attemptConst,
					//	delayIncrement: time.Second,
					//	log:            log,
					//	fnBuildReq:     buildReq,
					//},
				},
			},
		*/
		chErr: make(chan error),
	}
}

// Повторный вызов функции fnSend attempts раз.
// Возвращает ошибку с кол-вом вызовов функции и самой ошибкой fnSend.
func retry(ctx context.Context, attempts int, sleep time.Duration, fnSend func() error) error {
	var err error

	for i := 0; i <= attempts; i++ {
		if i > 0 {
			log.Printf("Повтор [%d] после ошибки: %v\n", i, err)

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
func (a *Agent) send(ctx context.Context, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.urlToSend, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("err build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	fmt.Printf("AGENT-KEY [%s]\n", a.cfg.key)
	if len(data) != 0 {
		sum, err := hash.SHA256(data, []byte(a.cfg.key))
		if err != nil {
			log.Printf("err %v\n", err)
			return err
		} else {
			req.Header.Set("HashSHA256", hex.EncodeToString(sum))
		}
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
func (a *Agent) Send(ctx context.Context, arr []model.Metric) error {
	data, err := json.Marshal(model.BuildArrMetricJSON(arr))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return retry(ctx, attemptConst, time.Second, func() error {
		return a.send(ctx, data)
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
	log.Printf(
		"start agent: addr[%s] poolInt[%d] reportInt[%d] key[%s] rateLimit[%d]\n",
		a.cfg.addr, a.cfg.pollInterval, a.cfg.reportInterval, a.cfg.key, a.cfg.rateLimit,
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

	fnSend := func(arr []model.Metric) error { return a.Send(ctxCan, arr) }
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
	name     string
	duration time.Duration
	fn       func() error
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
			defer wgTask.Done()

			for {
				select {
				case <-ctx.Done():
					chErr <- fmt.Errorf("задача [%v] отменена", task.name)
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

func buildReq(ctx context.Context, url string, data io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, data)
	if err != nil {
		return nil, err
	}

	return req, nil
}

/*
func buildRequest(ctx context.Context, url string, data []byte) (*http.Request, error) {
	fmt.Printf("LEN [%d]\n", len(data))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
*/

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
		// fmt.Println("wait fanIN OK")
	}()

	return finalCh
}

// Запуск worker'ов.
func runWorkerPool(rateLimit int, jobc <-chan []model.Metric, fnSend func([]model.Metric) error) <-chan error {
	var wgPool sync.WaitGroup

	errc := make(chan error)

	for idWorker := 0; idWorker < rateLimit; idWorker++ {
		wgPool.Add(1)

		go func(id int) {
			// fmt.Printf("start worker [%v]\n", id)
			//	defer fmt.Printf("stopt worker [%v]\n", id)

			for list := range jobc {
				if err := fnSend(list); err != nil {
					errc <- err
				}
			}

			wgPool.Done()
		}(idWorker)
	}

	go func() {
		wgPool.Wait()
		close(errc)
		// fmt.Println("waitPool OK")
	}()

	return errc
}
