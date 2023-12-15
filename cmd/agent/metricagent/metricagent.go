package metricagent

import (
	"fmt"
	"log"
	"net/http"
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
}

func New(opts ...FuncOpt) *MetricClient {
	agent := &MetricClient{
		stats:          *stats.NewStats(),
		store:          memstorage.New(memstorage.NewGaugeRepo(), memstorage.NewCounterRepo()),
		addr:           AddressDefault,
		pollInterval:   PollIntervalDefault,
		reportInterval: ReportIntervalDefault,
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
	err := c.UpdateMetrics()
	if err != nil {
		fmt.Printf("ERR-1 %v\n", err)
		return err
	}

	time.Sleep(time.Duration(c.pollInterval) * time.Second)

	err = c.SendMetrics()
	if err != nil {
		return err
	}
	time.Sleep(time.Duration(c.reportInterval) * time.Second)

	return nil
}

// UpdateMetrics Обновление всех метрик из пакета runtime и сохраниение в хранилище
func (c *MetricClient) UpdateMetrics() error {
	return c.stats.ReadToStore(c.store)
}

// SendMetrics Чтение и отправка всех сохраненых метрик
func (c *MetricClient) SendMetrics() error {
	for name, val := range c.store.GaugeRepo().List() {
		err := c.SendMetric(metric.GaugeType.String(), name, val)
		if err != nil {
			return err
		}
	}
	for name, val := range c.store.CounterRepo().List() {
		err := c.SendMetric(metric.CounterType.String(), name, val)
		if err != nil {
			return err
		}
	}

	return nil
}

// SendMetric Оправка метрики агентом на сервер по адресу
func (c *MetricClient) SendMetric(typeStr, name, valStr string) error {
	url := fmt.Sprintf(
		"http://%s/update/%s/%s/%s",
		c.addr, typeStr, name, valStr)
	client := &http.Client{}

	request, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "text/plain")
	response, err := client.Do(request)

	if err != nil {
		return err
	}
	//io.Copy(os.Stdout, response.Body)
	defer response.Body.Close()

	return nil
}

func (c *MetricClient) SendMetric1(typeStr, name, valStr string) error {
	url := fmt.Sprintf(
		"http://%s/update/%s/%s/%s",
		c.addr, typeStr, name, valStr)
	_ = url
	url2 := "http://localhost:8080/update/gauge/MyGauge/123"
	client := &http.Client{}
	resp, err := client.Post(url2, "text/plain", http.NoBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
