package metricagent

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AndreyVLZ/metrics/internal/metric"
	"github.com/AndreyVLZ/metrics/internal/stats"
	"github.com/AndreyVLZ/metrics/internal/storage"
	"github.com/AndreyVLZ/metrics/internal/storage/memstorage"
)

type MetricClient struct {
	stats          stats.Stats
	store          storage.Storage
	addr           string
	port           string
	contentType    string
	pollInterval   int
	reportInterval int
}

func New() *MetricClient {
	return &MetricClient{
		stats: *stats.NewStats(),
		store: memstorage.New(memstorage.NewGaugeRepo(), memstorage.NewCounterRepo()),
		//addr:           "localhost",
		addr:           "127.0.0.1",
		port:           "8080",
		contentType:    "text/plain",
		pollInterval:   2,
		reportInterval: 10,
	}
}

// AddMetric сохраняет в репозиторий значение произвольной метрики
func (c *MetricClient) AddMetric(name string, typeStr string, valStr string) error {
	return c.store.Set(name, typeStr, valStr)
}

// Start запускет агент
func (c *MetricClient) Start() error {
	/*
		var err error
		go func() {
			for err == nil {
				err = c.UpdateMetrics()
				time.Sleep(time.Duration(c.pollInterval) * time.Second)
			}
		}()

		go func() {
			for err == nil {
				err = c.SendMetrics()
				time.Sleep(time.Duration(c.reportInterval) * time.Second)
			}
		}()

		if err != nil {
			return err
		}
	*/
	//time.Sleep(30 * time.Second)

	err := c.UpdateMetrics()
	if err != nil {
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
	return c.stats.ReadToRepo(c.store.GaugeRepo())
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
		"http://%s:%s/update/%s/%s/%s",
		c.addr, c.port, typeStr, name, valStr)
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
