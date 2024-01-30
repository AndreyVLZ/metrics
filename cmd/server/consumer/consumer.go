package consumer

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/AndreyVLZ/metrics/internal/metric"
)

type Consumer struct {
	fileName string
	file     *os.File
	scanner  *bufio.Scanner
}

func New(filename string) *Consumer {
	return &Consumer{
		fileName: filename,
	}
}

func (c *Consumer) Open() error {
	file, err := os.OpenFile(c.fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	c.file = file
	c.scanner = bufio.NewScanner(file)
	return nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

func (c *Consumer) ReadMetric() ([]metric.MetricDB, error) {
	arr := []metric.MetricDB{}
	for c.scanner.Scan() {
		// читаем данные из scanner
		data := c.scanner.Bytes()

		var metricDB metric.MetricDB
		err := json.Unmarshal(data, &metricDB)
		if err != nil {
			return nil, err
		}

		arr = append(arr, metricDB)
	}

	return arr, nil
}
