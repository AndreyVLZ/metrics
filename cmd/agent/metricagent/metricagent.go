package metricagent

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

	"github.com/AndreyVLZ/metrics/cmd/agent/stats"
	"github.com/AndreyVLZ/metrics/cmd/server/route/mainhandler"
	"github.com/AndreyVLZ/metrics/internal/hash"
	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/AndreyVLZ/metrics/internal/storage"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

const (
	AddressDefault        = "localhost:8080"
	PollIntervalDefault   = 2
	ReportIntervalDefault = 10
)

const numJobs int = 2 // runtime && gopsutil

var ErrRetryCancel = errors.New("retry cancel")

type MetricClient struct {
	stats          stats.Stats
	store          storage.Storage
	client         *http.Client
	addr           string
	pollInterval   int
	reportInterval int
	key            string
	rateLimit      int
}

func New(opts ...FuncOpt) *MetricClient {
	agent := &MetricClient{
		stats:          *stats.NewStats(),
		store:          memstorage.New(),
		addr:           AddressDefault,
		pollInterval:   PollIntervalDefault,
		reportInterval: ReportIntervalDefault,
		client:         &http.Client{},
	}

	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

func workerPoll(ctx context.Context, rateLimit int, jobs <-chan []metric.MetricDB, errc chan<- error, fn func(context.Context, []metric.MetricDB) error) {
	wgCtx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup

	go func() {
		wg.Wait()
		log.Println("all workers stoped")
		cancel()
	}()

	for id := 1; id <= rateLimit; id++ {
		log.Printf("worker [%d] regist\n", id)
		wg.Add(1)
		go func(id int) {
			defer func() {
				wg.Done()
				log.Printf("worker [%d] stop\n", id)
			}()

			for batch := range jobs {
				if err := fn(wgCtx, batch); err != nil {
					if !errors.Is(err, ErrRetryCancel) {
						cancel()
						errc <- err
					}
					return
				}
			}
		}(id)
	}
}

type task struct {
	name    string
	duraton time.Duration
	fn      func() error
}

func (t *task) run(ctx context.Context) error {
	for {
		select {
		case <-time.After(t.duraton):
			err := t.fn()
			if err != nil {
				log.Printf("err %v[%v]\n", err, t.name)
				return err
			}
		case <-ctx.Done():
			log.Printf("exit %v\n", t.name)
			return nil
		}
	}
}

// Start запускет агент
func (c *MetricClient) Start() {
	log.Printf(
		"start agent: addr[%s] poolInt[%d] reportInt[%d] key[%s] rateLimit[%d]\n",
		c.addr, c.pollInterval, c.reportInterval, c.key, c.rateLimit,
	)

	mainCtx := context.Background()

	chBatch := make(chan []metric.MetricDB, 2)
	// Канал с ошибками для Task-ов и SendBatch
	errc := make(chan error, 1)

	// Создаем список задач для агента
	tasks := []task{
		{
			name:    "update",
			duraton: time.Duration(c.pollInterval) * time.Second,
			fn: func() error {
				return c.updateAllMetrics()
			},
		},

		{
			name:    "send runtime",
			duraton: time.Duration(c.reportInterval) * time.Second,
			fn: func() error {
				chBatch <- c.store.List(mainCtx)
				return nil
			},
		},

		{
			name:    "send gopsutil",
			duraton: time.Duration(c.reportInterval/2) * time.Second,
			fn: func() error {
				arr, err := goUtil()
				if err != nil {
					return err
				}
				chBatch <- arr
				return nil
			},
		},
	}

	ctx, cancel := context.WithCancel(mainCtx)
	defer cancel()

	// Запускаем все задачи
	var wg sync.WaitGroup
	for i := range tasks {
		wg.Add(1)
		go func(t task) {
			defer wg.Done()
			if err := t.run(ctx); err != nil {
				errc <- err
			}
		}(tasks[i])
	}

	workerPoll(ctx, c.rateLimit, chBatch, errc, c.SendBatch)

	// Отменяем все задачи если произошла ошибка
	if err := <-errc; err != nil {
		log.Printf("err %v\n", err)
		cancel()
	}

	func() {
		// Ждем когда завершатся все Task-и
		wg.Wait()
		// Закрываем все каналы
		close(chBatch)
		close(errc)
	}()
}

func (c *MetricClient) updateAllMetrics() error {
	if err := c.stats.ReadToStore(c.store); err != nil {
		return err
	}
	if err := c.randomValueUpdate(); err != nil {
		return err
	}

	return nil
}

func (c *MetricClient) randomValueUpdate() error {
	_, err := c.store.Set(context.Background(), metric.NewMetricDB("RandomValue", metric.Gauge(rand.Float64())))
	return err
}

func (c *MetricClient) SendBatch(ctx context.Context, metrics []metric.MetricDB) error {
	metricsJSON := make([]mainhandler.MetricsJSON, len(metrics))
	for i := range metrics {
		metricJSON, err := mainhandler.NewMetricJSONFromMetricDB(metrics[i])
		if err != nil {
			return err
		}

		metricsJSON[i] = metricJSON
	}

	metricsJSONBytes, err := json.Marshal(metricsJSON)
	if err != nil {
		return err
	}

	return retry(ctx, 3, time.Second, func() error {
		return c.sendData(ctx, "updates", metricsJSONBytes)
	})
}

func (c *MetricClient) sendData(ctx context.Context, endPoint string, data []byte) error {
	url := fmt.Sprintf("http://%s/%s/", c.addr, endPoint)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("err build request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if len(data) != 0 {
		sum, err := hash.SHA256(data, []byte(c.key))
		if err == nil {
			req.Header.Set("HashSHA256", hex.EncodeToString(sum))
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("err to send request: %s", err)
	}

	if err = resp.Body.Close(); err != nil {
		return fmt.Errorf("err body close: %s", err)
	}

	return nil
}

func retry(ctx context.Context, attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i <= attempts; i++ {
		if i > 0 {
			log.Printf("Повтор [%d] после ошибки %v\n", i, err)

			select {
			case <-ctx.Done():
				return ErrRetryCancel
			case <-time.After(sleep):
				sleep += 2 * time.Second
			}
		}

		err = f()
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("попыток %d, error: %s", attempts, err)
}

func goUtil() ([]metric.MetricDB, error) {
	cpuCount, err := cpu.Counts(true)
	if err != nil {
		return nil, err
	}

	vmStats, err := mem.VirtualMemory()
	_ = vmStats
	if err != nil {
		return nil, err
	}

	arr := make([]metric.MetricDB, 0, 3)
	arr = append(arr,
		metric.NewMetricDB("TotalMemory", metric.Gauge(float64(vmStats.Available))),
		metric.NewMetricDB("FreeMemory", metric.Gauge(float64(vmStats.Free))),
		metric.NewMetricDB("CPUutilization1", metric.Gauge(float64(cpuCount))),
	)

	return arr, nil
}
