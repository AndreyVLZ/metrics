package filestore

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/AndreyVLZ/metrics/internal/model"
)

// Структура метрик для хранения в файле.
type fileMetric struct {
	NameID string     `json:"mName"`
	TypeID model.Type `json:"mType"`
	Val    *float64   `json:"mVal,omitempty"`
	Delta  *int64     `json:"mDelta,omitempty"`
}

func (fm fileMetric) buildModelMetric() model.Metric {
	return model.NewMetric(
		model.Info{MName: fm.NameID, MType: fm.TypeID},
		model.Value{Delta: fm.Delta, Val: fm.Val},
	)
}

func buildFileMetric(met model.Metric) fileMetric {
	return fileMetric{
		NameID: met.MName,
		TypeID: met.MType,
		Val:    met.Val,
		Delta:  met.Delta,
	}
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
	if err := f.consumer.Open(); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := f.producer.Open(); err != nil {
		return fmt.Errorf("%w", err)
	}

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

func (f *File) WriteMetric(met model.Metric) error {
	data, err := json.Marshal(buildFileMetric(met))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := f.producer.Write(data); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (f *File) WriteBatch(arr []model.Metric) error {
	if err := f.producer.Trunc(); err != nil {
		return fmt.Errorf("%w", err)
	}

	for i := range arr {
		if err := f.WriteMetric(arr[i]); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

func (f *File) ReadBatch() ([]model.Metric, error) {
	var fileMet fileMetric

	arr := make([]model.Metric, 0)

	for f.consumer.Scan() {
		fileMet = fileMetric{}
		data := f.consumer.Bytes()

		err := json.Unmarshal(data, &fileMet)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		arr = append(arr, fileMet.buildModelMetric())
	}

	if err := f.consumer.Err(); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return arr, nil
}
