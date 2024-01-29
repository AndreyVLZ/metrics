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

/*
func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}
*/

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

	// NOTE:
	//return reverseArray(arr), nil
	return arr, nil
}

func reverseArray(arr []metric.MetricDB) []metric.MetricDB {
	l := len(arr)
	for i := 0; i < l/2; i++ {
		arr[i], arr[l-1-i] = arr[l-1-i], arr[i]
	}

	return arr
}

/*
func corectiveArr(arr []metric.MetricDB) []metric.MetricDB {
	arrToSend := []metric.MetricDB{}

	mapNames := make(map[string]struct{}, len(arr))

	for i := len(arr) - 1; i >= 0; i-- {
		_, ok := mapNames[arr[i].Name()]
		if !ok {
			arrToSend = append(arrToSend, arr[i])
		}
		mapNames[arr[i].Name()] = struct{}{}
	}

	return arrToSend
}
*/
