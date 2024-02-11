package producer

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/AndreyVLZ/metrics/internal/metric"
)

type Producer struct {
	fileName string
	file     *os.File
	writer   *bufio.Writer
}

func New(fileName string) *Producer {
	return &Producer{
		fileName: fileName,
	}
}

func (p *Producer) Open() error {
	file, err := os.OpenFile(p.fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	p.file = file
	p.writer = bufio.NewWriter(file)

	return nil
}

/*
func (p *Producer) Trunc1() error {
	if p.file != nil {
		if err := p.Close(); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(p.fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	p.file = file
	p.writer = bufio.NewWriter(file)
	return nil
}
*/

func (p *Producer) Trunc() error {
	return p.file.Truncate(0)
}

func (p *Producer) WriteMetric(metricDB *metric.MetricDB) error {
	data, err := json.Marshal(&metricDB)
	if err != nil {
		return err
	}

	// записываем событие в буфер
	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return p.writer.Flush()
	//return nil
}

func (p *Producer) Close() error {
	return p.file.Close()
}
