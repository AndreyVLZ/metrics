package memstorage

import (
	"errors"
	"strconv"
)

const (
	gaugeConst = "gauge"
	countConst = "counter"
)

type MemStorage struct {
	gaugeStore map[string]float64
	countStore map[string]int64
}

func New() *MemStorage {
	return &MemStorage{
		gaugeStore: make(map[string]float64),
		countStore: make(map[string]int64),
	}
}

func (s *MemStorage) Set(typeStr, name, val string) error {
	var err error
	switch typeStr {
	case gaugeConst:
		err = s.setGauge(name, val)
	case countConst:
		err = s.setCount(name, val)
	default:
		return errors.New("not support type")
	}

	if err != nil {
		return err
	}

	return nil
}

func (s *MemStorage) setGauge(name, val string) error {
	newFloat, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return err
	}

	s.gaugeStore[name] = newFloat

	return nil
}

func (s *MemStorage) setCount(name, val string) error {
	newInt, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return err
	}

	_, ok := s.countStore[name]
	if !ok {
		s.countStore[name] = 0
	}

	s.countStore[name] += newInt

	return nil
}
