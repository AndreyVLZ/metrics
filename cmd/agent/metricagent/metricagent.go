package metricagent

import (
	"bytes"
	"context"
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
	exit           chan (struct{})
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
	_, err := c.store.Set(context.Background(), metric.NewMetricDB("RandomValue", metric.Gauge(rand.Float64())))
	return err
}

// SendMetrics Чтение и отправка всех сохраненых метрик
func (c *MetricClient) SendMetrics(wg *sync.WaitGroup) {
	var err error

	for {
		select {
		case <-time.After(time.Duration(c.reportInterval) * time.Second):
			metrics := c.store.List(context.Background())
			err = c.SendBatch(metrics)
			if err != nil {
				log.Printf("err send batch metrics> %v\n", err)
				c.exit <- struct{}{}
			}
		case <-c.exit:
			wg.Done()
		}
	}
}

func (c *MetricClient) SendBatch(metrics []metric.MetricDB) error {
	metricsJSON := make([]mainhandler.MetricsJSON, len(metrics))
	for i := range metrics {
		metricJSON, err := mainhandler.NewMetricJSONFromMetricDB(metrics[i])
		if err != nil {
			return err
		}

		metricsJSON[i] = metricJSON
	}

	JSON, err := json.Marshal(metricsJSON)
	if err != nil {
		return err
	}

	return retry(4, time.Second, func() error {
		return c.sendData(JSON)
	})
}

func (c *MetricClient) sendData(data []byte) error {
	url := fmt.Sprintf("http://%s/updates/", c.addr)
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("err build request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("err to send request: %s", err)
	}

	if err = resp.Body.Close(); err != nil {
		return fmt.Errorf("err body close: %s", err)
	}

	return nil
}

func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			log.Printf("Повтор псле ошибки %v\n", err)
			fmt.Printf("SLEEP %v\n", sleep)
			time.Sleep(sleep)
			sleep += 2 * time.Second
		}

		err = f()
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("попыток %d, error: %s", attempts, err)
}

func (c *MetricClient) SendMetricPost(metric metric.MetricDB) error {

	metricJSON, err := mainhandler.NewMetricJSONFromMetricDB(metric)
	if err != nil {
		return err
	}

	data, err := json.Marshal(metricJSON)
	if err != nil {
		return err
	}
	return c.sendData(data)
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
