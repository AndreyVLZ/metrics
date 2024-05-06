package filestore

import (
	"bufio"
	"encoding/json"
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/AndreyVLZ/metrics/internal/model"
)

// Структура метрик для хранения в файле.
type metric struct {
	Name  string   `json:"mName"`
	MType string   `json:"mType"`
	Val   *float64 `json:"mVal,omitempty"`
	Delta *int64   `json:"mDelta,omitempty"`
}

func buildMetric[VT model.ValueType](met model.MetricRepo[VT]) metric {
	m := metric{
		Name:  met.Name(),
		MType: met.Type(),
	}

	switch met.Type() {
	case model.TypeCountConst.String():
		vv := int64(met.Value())
		m.Delta = &vv
	case model.TypeGaugeConst.String():
		vv := float64(met.Value())
		m.Val = &vv
	}

	return m
}

type File struct {
	filePath string
	producer *Producer
	consumer *Consumer
}

type Consumer struct {
	filePath string
	file     *os.File
	scanner  *bufio.Scanner
}

func NewConsumer(filePath string) *Consumer {
	return &Consumer{filePath: filePath}
}

func (c *Consumer) Open() error {
	file, err := os.OpenFile(c.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	c.file = file
	c.scanner = bufio.NewScanner(file)

	return nil
}

func (c *Consumer) Scan() bool    { return c.scanner.Scan() }
func (c *Consumer) Bytes() []byte { return c.scanner.Bytes() }
func (c *Consumer) Err() error    { return c.scanner.Err() }
func (c *Consumer) Close() error  { return c.file.Close() }

type Producer struct {
	filePath string
	file     *os.File
	writer   *bufio.Writer
}

func NewProducer(filePath string) *Producer {
	return &Producer{filePath: filePath}
}

func (p *Producer) Open() error {
	file, err := os.OpenFile(p.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	p.file = file
	p.writer = bufio.NewWriter(file)

	return nil
}

func (p *Producer) Write(data []byte) error {
	// записываем событие в буфер
	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	return p.writer.Flush()
}

func (p *Producer) Close() error { return p.file.Close() }
func (p *Producer) Trunc() error { return p.file.Truncate(0) }

func NewFile(filePath string) *File {
	return &File{
		filePath: filePath,
		consumer: NewConsumer(filePath),
		producer: NewProducer(filePath),
	}
}

func (f *File) Open() error {
	/*
		file, err := os.OpenFile(f.filePath, os.O_RDWR|os.O_APPEND, 0777)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	*/

	if err := f.consumer.Open(); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := f.producer.Open(); err != nil {
		return fmt.Errorf("%w", err)
	}

	// f.file = file
	// f.writer = bufio.NewWriter(file)
	// f.scanner = bufio.NewScanner(file)

	return nil
}

func (f *File) Close() error {
	if err := f.consumer.Close(); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := f.producer.Close(); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

//func (f *File) Trunc() error { return f.file.Truncate(0) }

func (f *File) WriteMetric(met metric) error {
	data, err := json.Marshal(met)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := f.producer.Write(data); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (f *File) ReadBatch() (model.Batch, error) {
	clist := make([]model.MetricRepo[int64], 0)
	glist := make([]model.MetricRepo[float64], 0)

	for f.consumer.Scan() {
		data := f.consumer.Bytes()

		var metricFile metric

		err := json.Unmarshal(data, &metricFile)
		if err != nil {
			return model.Batch{}, fmt.Errorf("%w", err)
		}

		switch metricFile.MType {
		case model.TypeCountConst.String():
			clist = append(clist, model.NewMetricRepo(metricFile.Name, model.TypeCountConst, *metricFile.Delta))
		case model.TypeGaugeConst.String():
			glist = append(glist, model.NewMetricRepo(metricFile.Name, model.TypeGaugeConst, *metricFile.Val))
		}
	}

	if err := f.consumer.Err(); err != nil {
		return model.Batch{}, fmt.Errorf("%w", err)
	}

	return model.Batch{
		CList: clist,
		GList: glist,
	}, nil
}

func (f *File) WriteBatch(batch model.Batch) error {
	if err := f.producer.Trunc(); err != nil {
		return fmt.Errorf("%w", err)
	}

	for i := range batch.CList {
		metL := batch.CList[i]
		val := metL.Value()

		met := metric{
			Name:  metL.Name(),
			MType: metL.Type(),
			Delta: &val,
		}

		if err := f.WriteMetric(met); err != nil {
			return fmt.Errorf("writeBatch: %w", err)
		}
	}

	for i := range batch.GList {
		metL := batch.GList[i]
		val := metL.Value()
		met := metric{
			Name:  metL.Name(),
			MType: metL.Type(),
			Val:   &val,
		}

		if err := f.WriteMetric(met); err != nil {
			return fmt.Errorf("writeBatch: %w", err)
		}
	}

	return nil
}
