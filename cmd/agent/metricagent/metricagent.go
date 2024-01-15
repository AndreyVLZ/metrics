package metricagent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/AndreyVLZ/metrics/cmd/agent/stats"
	"github.com/AndreyVLZ/metrics/cmd/server/route/mainhandler"
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
		err = c.updateAllMetrics()
		time.Sleep(time.Duration(c.pollInterval) * time.Second)
	}
	wg.Done()

	log.Printf("err> %v\n", err)
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
	return c.store.Set(metric.NewMetricDB("RandomValue", metric.Gauge(rand.Float64())))
}

// SendMetrics Чтение и отправка всех сохраненых метрик
func (c *MetricClient) SendMetrics(wg *sync.WaitGroup) {
	var err error

	for err == nil {
		time.Sleep(time.Duration(c.reportInterval) * time.Second)
		metrics := c.store.List()
		for i := range metrics {
			err = c.SendMetricPost(metrics[i])
			if err != nil {
				log.Printf("err> %v\n", err)
			}
		}
	}

	wg.Done()
}

func (c *MetricClient) SendMetricPost(metric metric.MetricDB) error {
	url := fmt.Sprintf("http://%s/update/", c.addr)

	metricJSON, err := mainhandler.NewMetricJSONFromMetricDB(metric)
	if err != nil {
		return err
	}

	data, err := json.Marshal(metricJSON)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		log.Printf("err build request %v\n", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("err to send request: %s", err)
	} else {
		if err = resp.Body.Close(); err != nil {
			log.Printf("err body close: %s", err)
		}
	}

	return nil
}

// SendMetric Оправка метрики агентом на сервер по адресу
func (c *MetricClient) SendMetric(metric metric.MetricDB) error {
	url := fmt.Sprintf(
		"http://%s/update/%s/%s/%s",
		c.addr, metric.Type(), metric.Name(), metric.Valuer.String())

	res, err := c.client.Post(url, "text/plain", http.NoBody)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}
