package field

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/AndreyVLZ/metrics/pkg/parser/perr"
)

var (
	errContertToString = errors.New("convert to string")
	errConvertToBool   = errors.New("convert to bool")
)

var myField *field

// init Инициализация глобальных переменных.
func init() {
	var once sync.Once

	once.Do(
		func() {
			myField = &field{
				data: make(map[string]interface{}, 0),
			}
		},
	)
}

type field struct {
	data map[string]interface{}
}

func (f *field) unmarshal(fileByte []byte) error {
	if err := json.Unmarshal(fileByte, &f.data); err != nil {
		return fmt.Errorf("field unmarshal: %w", err)
	}

	return nil
}

func (f *field) duration(fieldName string) func(*time.Duration) error {
	return func(valDur *time.Duration) error {
		jsonVal, isExist := f.data[fieldName]
		if !isExist {
			return perr.ErrNotSet
		}

		jsonStr, isOK := jsonVal.(string)
		if !isOK {
			return errContertToString
		}

		dur, err := time.ParseDuration(jsonStr)
		if err != nil {
			return fmt.Errorf("parse [%s] to duration: %w", jsonStr, err)
		}

		*valDur = dur

		return nil
	}
}

func (f *field) bool(fieldName string) func(*bool) error {
	return func(valBool *bool) error {
		jsonVal, isExist := f.data[fieldName]
		if !isExist {
			return perr.ErrNotSet
		}

		jsonBool, isOK := jsonVal.(bool)
		if !isOK {
			return errConvertToBool
		}

		*valBool = jsonBool

		return nil
	}
}

func (f *field) string(fieldName string) func(*string) error {
	return func(str *string) error {
		jsonVal, isExist := f.data[fieldName]
		if !isExist {
			return perr.ErrNotSet
		}

		jsonStr, isOK := jsonVal.(string)
		if !isOK {
			return errContertToString
		}

		*str = jsonStr

		return nil
	}
}

// Unmarshal ...
func Unmarshal(fileByte []byte) error { return myField.unmarshal(fileByte) }

// Duration Читает field как time.Duration. Возвращает функцию установки time.Duration-значения.
func Duration(fieldName string) func(*time.Duration) error {
	return myField.duration(fieldName)
}

// Bool Читает field как bool. Возвращает функцию установки bool-значения.
func Bool(fieldName string) func(*bool) error {
	return myField.bool(fieldName)
}

// String Читает field как string. Возвращает функцию установки string-значения.
func String(fieldName string) func(*string) error {
	return myField.string(fieldName)
}
