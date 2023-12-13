package metricagent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AndreyVLZ/metrics/cmd/agent/config"
	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/AndreyVLZ/metrics/internal/stats"
	"github.com/AndreyVLZ/metrics/internal/storage"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
)

type MetricClient struct {
	stats       stats.Stats
	store       storage.Storage
	contentType string
	conf        *config.Config
}

func New(conf *config.Config) *MetricClient {
	return &MetricClient{
		stats:       *stats.NewStats(),
		store:       memstorage.New(memstorage.NewGaugeRepo(), memstorage.NewCounterRepo()),
		conf:        conf,
		contentType: "text/plain",
	}
}

// AddMetric сохраняет в репозиторий значение произвольной метрики
func (c *MetricClient) AddMetric(name string, typeStr string, valStr string) error {
	return c.store.Set(name, typeStr, valStr)
}

// Start запускет агент
func (c *MetricClient) Start() error {
	err := c.UpdateMetrics()
	if err != nil {
		return err
	}
	time.Sleep(time.Duration(c.conf.PollInterval) * time.Second)

	err = c.SendMetrics()
	if err != nil {
		return err
	}
	time.Sleep(time.Duration(c.conf.ReportInterval) * time.Second)

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

	return nil
}

// SendMetric Оправка метрики агентом на сервер по адресу [addr:port]
func (c *MetricClient) SendMetric(typeStr, name, valStr string) error {
	url := fmt.Sprintf(
		"http://%s/update/%s/%s/%s",
		c.conf.Addr, typeStr, name, valStr)
	client := &http.Client{}

	request, err := http.NewRequest(http.MethodPost, url, http.NoBody)
	if err != nil {
		return err
	}

	//c.setContentType(request)
	request.Header.Set("Content-Type", c.contentType)
	response, err := client.Do(request)

	if err != nil {
		return err
	}
	//io.Copy(os.Stdout, response.Body)
	response.Body.Close()

	return nil
}

// setContentType Установка заголовка "Content-Type" для текущего запроса
func (c *MetricClient) setContentType(req *http.Request) {
	req.Header.Set("Content-Type", c.contentType)
}
