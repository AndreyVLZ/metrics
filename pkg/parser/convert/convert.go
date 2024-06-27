package convert

import (
	"errors"
	"time"

	"github.com/AndreyVLZ/metrics/pkg/parser/perr"
)

// IntToDuration Конвертирует значение типа int, прочитаного из parsers, в time.Duration.
// Возвращает функцию установки time.Duration-значения.
func IntToDuration(duration time.Duration, parsers ...func(valInt *int) error) func(*time.Duration) error {
	return func(valDur *time.Duration) error {
		var (
			valInt int
			isSet  bool
		)

		for i := range parsers {
			err := parsers[i](&valInt)
			if err != nil {
				if !errors.Is(err, perr.ErrNotSet) {
					return err
				}

				continue
			}

			isSet = true
		}

		if !isSet {
			return perr.ErrNotSet
		}

		*valDur = time.Duration(valInt) * duration

		return nil
	}
}
