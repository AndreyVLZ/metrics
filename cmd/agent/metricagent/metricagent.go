package metricagent

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
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

type MetricClient struct {
	stats          stats.Stats
	store          storage.Storage
	client         *http.Client
	addr           string
	pollInterval   int
	reportInterval int
	key            string
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
	log.Printf(
		"start agent: addr[%s] poolInt[%d] reportInt[%d] key[%s]\n",
		c.addr, c.pollInterval, c.reportInterval, c.key,
	)
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
		<-time.After(time.Duration(c.reportInterval) * time.Second)
		metrics := c.store.List(context.Background())
		//c.Send2(metrics)
		err = c.SendBatch(metrics)
		if err != nil {
			log.Printf("err send batch metrics> %v\n", err)
			break
		}
	}
	wg.Done()
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

	return retry(3, time.Second, func() error {
		return c.sendData("updates", JSON)
	})
}

func hash(data []byte, key []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write(data)
	if err != nil {
		return nil, errors.New("err hash")
	}

	sum := h.Sum(nil)
	return sum, nil
}

func (c *MetricClient) sendData(endPoint string, data []byte) error {
	url := fmt.Sprintf("http://%s/%s/", c.addr, endPoint)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("err build request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if len(data) != 0 {
		sum, err := hash(data, []byte(c.key))
		if err == nil {
			_ = sum
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

func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i <= attempts; i++ {
		if i > 0 {
			log.Printf("Повтор [%d] после ошибки %v\n", i, err)
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

func (c *MetricClient) Send2(metrics []metric.MetricDB) {
	for i := range metrics {
		err := c.SendMetricPost(metrics[i])
		if err != nil {
			fmt.Printf("errSendOne %v\n", err)
		}
	}
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
	return c.sendData("update", data)
}

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

func SetKey(key string) FuncOpt {
	return func(c *MetricClient) {
		c.key = key
	}
}
