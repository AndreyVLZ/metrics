package agent

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/AndreyVLZ/metrics/agent/stats"
	"github.com/AndreyVLZ/metrics/internal/hash"
	"github.com/AndreyVLZ/metrics/internal/model"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

const (
	attemptConst          = 3
	durationTasckConst    = 2
	retryTimeoutStepConst = 2
)

type storage interface {
	List(ctx context.Context) (model.Batch, error)
	AddBatch(ctx context.Context, batch model.Batch) error
	Stop(ctx context.Context) error
	Start(ctx context.Context) error
}

type Agent struct {
	cfg    *Config
	stats  stats.Stats
	client *http.Client
	store  storage
	isStop bool
	exit   chan struct{}
}

func New(cfg *Config, store storage) *Agent {
	var httpClient http.Client
	return &Agent{
		cfg:    cfg,
		stats:  *stats.New(),
		client: &httpClient,
		store:  store,
		isStop: false,
		exit:   make(chan struct{}),
	}
}

func (a *Agent) Stop(ctx context.Context) error {
	if !a.isStop {
		<-a.exit
	}

	if err := a.store.Stop(ctx); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// Start запускает агент.
func (a *Agent) Start(ctx context.Context) error {
	log.Printf(
		"start agent: addr[%s] poolInt[%d] reportInt[%d] key[%s] rateLimit[%d]\n",
		a.cfg.addr, a.cfg.pollInterval, a.cfg.reportInterval, a.cfg.key, a.cfg.rateLimit,
	)

	if err := a.store.Start(ctx); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := a.start(ctx); err != nil {
		a.isStop = true
		return fmt.Errorf("%w", err)
	}

	a.exit <- struct{}{}

	return nil
}

/*
type taskPool struct {
	tasks []task
}

func (tp *taskPool) run(ctx context.Context) (<-chan model.Batch, <-chan error) {
	chBatch := make(chan model.Batch)
	errc := make(chan error)
	ctxTask, cancel := context.WithCancel(ctx)

	go func() {
		var wgTask sync.WaitGroup

		for i := range tp.tasks {
			wgTask.Add(1)

			go func(task task) {
				defer func() {
					wgTask.Done()
					fmt.Printf("exit %s\n", task.name)
					cancel()
				}()

				for {
					select {
					case <-ctxTask.Done():
						return
					case <-time.After(task.duration):
						if err := task.fn(ctxTask, chBatch); err != nil {
							errc <- err
							return
						}
					}
				}
			}(tp.tasks[i])
		}

		wgTask.Wait()
		fmt.Println("wgTask OK")
		close(chBatch)
		close(errc)
	}()

	return chBatch, errc
}

type workerPool struct {
	rateLimit int
}

func (wp *workerPool) run(ctx context.Context, a *Agent, chBatch <-chan model.Batch) <-chan error {
	var wgPool sync.WaitGroup

	ctxWPoll, cancel := context.WithCancel(ctx)
	errc := make(chan error)

	go func() {
		for i := 0; i < wp.rateLimit; i++ {
			wgPool.Add(1)

			go func() {
				for batch := range chBatch {
					if err := a.SendBatchByRetry(ctxWPoll, batch); err != nil {
						errc <- err

						cancel()
					}
					fmt.Println("FLUSH")
				}

				wgPool.Done()
				fmt.Println("exit worker")
			}()
		}
		wgPool.Wait()
		close(errc)
	}()

	return errc
}
func (a *Agent) start1(ctx context.Context) error {
	taskPool := taskPool{
		tasks: []task{
			{
				name:     "update", // сбор метрик (опрос runtime)
				duration: time.Duration(a.cfg.pollInterval) * time.Second,
				fn: func(ct context.Context, _ chan<- model.Batch) error {
					return a.updateAllMetrics(ct)
				},
			},

			{
				name:     "send runtime", // отправка сохраненых метрик
				duration: time.Duration(a.cfg.reportInterval) * time.Second,
				fn: func(ct context.Context, chBatch chan<- model.Batch) error {
					batch, err := a.store.List(ct)
					if err != nil {
						return fmt.Errorf("task sendRunTime: %w", err)
					}

					if len(batch.CList)+len(batch.GList) == 0 {
						return nil
					}

					chBatch <- batch

					return nil
				},
			},

			{
				name:     "send gopsutil", // сбор метрики из пакета gopsutil
				duration: time.Duration(a.cfg.reportInterval/durationTasckConst) * time.Second,
				fn: func(_ context.Context, chBatch chan<- model.Batch) error {
					batch, err := goUtil()
					fmt.Println(1)
					if err != nil {
						return fmt.Errorf("task sendGopsutil: %w", err)
					}
					fmt.Println(2)

					chBatch <- batch
					fmt.Println(3)

					return nil
				},
			},
		},
	}

	workerPool := workerPool{rateLimit: a.cfg.rateLimit}

	chBatch, errChTask := taskPool.run(ctx)
	errChWP := workerPool.run(ctx, a, chBatch)
	errc := make(chan error)

	go func(arrChErr []<-chan error) {
		var wgErr sync.WaitGroup

		for _, ch := range arrChErr {
			chClosure := ch

			wgErr.Add(1)

			go func() {
				for data := range chClosure {
					errc <- data
				}

				wgErr.Done()
			}()
		}

		wgErr.Wait()
		close(errc)
	}([]<-chan error{errChTask, errChWP})

	arrErr := make([]error, 0)
	for err := range errc {
		arrErr = append(arrErr, err)
	}

	return errors.Join(arrErr...)
}
*/

func (a *Agent) start(ctx context.Context) error {
	chBatch := make(chan model.Batch)
	errc := make(chan error)

	// Создаем список задач для агента
	tasks := []task{
		{
			name:     "update", // сбор метрик (опрос runtime)
			duration: time.Duration(a.cfg.pollInterval) * time.Second,
			fn: func(ct context.Context) error {
				return a.updateAllMetrics(ct)
			},
		},

		{
			name:     "send runtime", // отправка сохраненых метрик
			duration: time.Duration(a.cfg.reportInterval) * time.Second,
			fn: func(ct context.Context) error {
				batch, err := a.store.List(ct)
				if err != nil {
					return fmt.Errorf("task sendRunTime: %w", err)
				}

				if len(batch.CList)+len(batch.GList) == 0 {
					return nil
				}

				chBatch <- batch

				return nil
			},
		},

		{
			name:     "send gopsutil", // сбор метрики из пакета gopsutil
			duration: time.Duration(a.cfg.reportInterval/durationTasckConst) * time.Second,
			fn: func(_ context.Context) error {
				batch, err := goUtil()
				if err != nil {
					return fmt.Errorf("task sendGopsutil: %w", err)
				}

				chBatch <- batch

				return nil
			},
		},
	}

	// Запускает все Task'и и worker'ы
	// Ждет их завершения
	go func() {
		var (
			wgTask sync.WaitGroup
			wgPool sync.WaitGroup
		)

		ctxCan, cancel := context.WithCancel(ctx)
		defer cancel()
		// Запуск Task'ов.
		func() {
			for i := range tasks {
				wgTask.Add(1)

				go func(task task) {
					defer func() {
						wgTask.Done()
						cancel() // Отменяем все Task'и и worker'ы, если произошла ошибка
					}()

					if err := task.run(ctxCan); err != nil {
						errc <- err

						return
					}
				}(tasks[i])
			}
		}()

		// Запуск worker'ов.
		func() {
			for i := 0; i < a.cfg.rateLimit; i++ {
				wgPool.Add(1)

				go func() {
					for batch := range chBatch {
						if err := a.SendBatchByRetry(ctxCan, batch); err != nil {
							errc <- err

							cancel() // При ошибке отпраляем сигнал на закрытие всех писателей [Task] в канал [chBatch]
						}
					}

					wgPool.Done()
				}()
			}
		}()

		wgTask.Wait()  // Ждем завершения работы всех задач [Task]
		close(chBatch) // Закрываем канал в который писали Task'и
		wgPool.Wait()  // Ждем завершения всех worker'ов
		close(errc)    // Закрываем канал к который писали worker'ы
	}()

	arrErr := []error{}
	for err := range errc {
		arrErr = append(arrErr, err)
	}

	return errors.Join(arrErr...)
}

// Запуск задачи.
func (t *task) run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(t.duration):
			if err := t.fn(ctx); err != nil {
				return fmt.Errorf("task[%s] by err %w", t.name, err)
			}
		}
	}
}

type task struct {
	name     string
	duration time.Duration
	// fn       func(context.Context, chan<- model.Batch) error
	fn func(context.Context) error
}

// Отправка метрик с фукцией повтора при ошибке.
func (a *Agent) SendBatchByRetry(ctx context.Context, batch model.Batch) error {
	metricsJSONBytes, err := json.Marshal(batch.ToArrMetricJSON())
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return retry(ctx, attemptConst, time.Second, func() error {
		return a.sendData(ctx, "updates", metricsJSONBytes)
	})
}

// Отправка метрик.
func (a *Agent) sendData(ctx context.Context, endPoint string, data []byte) error {
	url := fmt.Sprintf("http://%s/%s/", a.cfg.addr, endPoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("err build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if len(data) != 0 {
		sum, err := hash.SHA256(data, []byte(a.cfg.key))
		if err != nil {
			log.Printf("err %v\n", err)
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

func (a *Agent) updateAllMetrics(ctx context.Context) error {
	batch, err := a.stats.BuildBatch()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	batch.GList = append(batch.GList, model.NewMetricRepo("RandomValue", model.TypeGaugeConst, rand.Float64()))

	if err := a.store.AddBatch(ctx, batch); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func retry(ctx context.Context, attempts int, sleep time.Duration, f func() error) error {
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

		err = f()
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("попыток %d, error: %w", attempts, err)
}

func goUtil() (model.Batch, error) {
	cpuCount, err := cpu.Counts(true)
	if err != nil {
		return model.Batch{}, fmt.Errorf("%w", err)
	}

	vmStats, err := mem.VirtualMemory()
	if err != nil {
		return model.Batch{}, fmt.Errorf("%w", err)
	}

	gList := []model.MetricRepo[float64]{
		model.NewMetricRepo("TotalMemory", model.TypeGaugeConst, float64(vmStats.Available)),
		model.NewMetricRepo("FreeMemory", model.TypeGaugeConst, float64(vmStats.Free)),
		model.NewMetricRepo("CPUutilization1", model.TypeGaugeConst, float64(cpuCount)),
	}

	return model.Batch{GList: gList}, nil
}
