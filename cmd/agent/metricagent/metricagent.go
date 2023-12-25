package metricagent

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/AndreyVLZ/metrics/cmd/agent/stats"
	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/AndreyVLZ/metrics/internal/storage"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
)

const (
	AddressDefault        = "localhost:8080"
	PollIntervalDefault   = 2
	ReportIntervalDefault = 10
)

type FuncOpt func(c *MetricClient)

func SetAddr(addr string) FuncOpt {
	return func(c *MetricClient) {
		c.addr = addr
	}
}

func SetPollInterval(pollInterval int) FuncOpt {
	return func(c *MetricClient) {
		c.pollInterval = pollInterval
	}
}

func SetReportInterval(reportInterval int) FuncOpt {
	return func(c *MetricClient) {
		c.reportInterval = reportInterval
	}
}

type MetricClient struct {
	stats          stats.Stats
	store          storage.Storage
	addr           string
	pollInterval   int
	reportInterval int
	client         *http.Client
}

func New(opts ...FuncOpt) *MetricClient {
	agent := &MetricClient{
		stats:          *stats.NewStats(),
		store:          memstorage.New(memstorage.NewGaugeRepo(), memstorage.NewCounterRepo()),
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

// AddMetric сохраняет в репозиторий значение произвольной метрики
func (c *MetricClient) AddMetric(typeStr string, name string, valStr string) error {
	return c.store.Set(typeStr, name, valStr)
}

// Start запускет агент
func (c *MetricClient) Start() error {
	log.Printf("start agent: %s %v %v\n", c.addr, c.pollInterval, c.reportInterval)
	var wg sync.WaitGroup

	wg.Add(1)
	go c.UpdateMetrics(&wg)
	wg.Add(1)
	go c.SendMetrics(&wg)

	wg.Wait()

	return nil
}

// UpdateMetrics Обновление всех метрик из пакета runtime и сохраниение в хранилище
func (c *MetricClient) UpdateMetrics(wg *sync.WaitGroup) {
	var err error

	for err == nil {
		time.Sleep(time.Duration(c.pollInterval) * time.Second)
		err = c.stats.ReadToStore(c.store)
	}
	wg.Done()
	log.Printf("err> %v\n", err)
}

// SendMetrics Чтение и отправка всех сохраненых метрик
func (c *MetricClient) SendMetrics(wg *sync.WaitGroup) {
	var err error

	for err == nil {
		time.Sleep(time.Duration(c.reportInterval) * time.Second)
		for name, val := range c.store.GaugeRepo().List() {
			err = c.SendMetric(metric.GaugeType.String(), name, val)
			if err != nil {
				log.Printf("err> %v\n", err)
			}
		}
		for name, val := range c.store.CounterRepo().List() {
			err = c.SendMetric(metric.CounterType.String(), name, val)
			if err != nil {
				log.Printf("err> %v\n", err)
			}
		}
	}
	wg.Done()
}

// SendMetric Оправка метрики агентом на сервер по адресу
func (c *MetricClient) SendMetric(typeStr, name, valStr string) error {
	url := fmt.Sprintf(
		"http://%s/update/%s/%s/%s",
		c.addr, typeStr, name, valStr)

	res, err := c.client.Post(url, "text/plain", http.NoBody)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
