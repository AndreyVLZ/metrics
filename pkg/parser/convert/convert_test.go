package convert

import (
	"errors"
	"testing"
	"time"

	"github.com/AndreyVLZ/metrics/pkg/parser/perr"
	"github.com/stretchr/testify/assert"
)

func TestInToDuration(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		valInt := 777

		dur := time.Second

		fnConvert := IntToDuration(dur,
			func(intPtr *int) error {
				val := 999
				*intPtr = val

				return nil
			},
			func(intPtr *int) error {
				*intPtr = valInt

				return nil
			},
			func(_ *int) error { return perr.ErrNotSet },
		)

		var acVal time.Duration
		if err := fnConvert(&acVal); err != nil {
			t.Error(err)
		}

		assert.Equal(t, time.Duration(valInt)*dur, acVal)
	})

	t.Run("err custom", func(t *testing.T) {
		valInt := 777
		myErr := errors.New("custom err")

		dur := time.Second

		fnConvert := IntToDuration(dur,
			func(_ *int) error { return myErr },
			func(intPtr *int) error {
				*intPtr = valInt

				return nil
			},
		)

		var acVal time.Duration
		err := fnConvert(&acVal)

		assert.Equal(t, myErr, err)
	})

	t.Run("err not set", func(t *testing.T) {
		dur := time.Second

		fnConvert := IntToDuration(dur,
			func(_ *int) error { return perr.ErrNotSet },
		)

		var acVal time.Duration
		err := fnConvert(&acVal)

		assert.Equal(t, perr.ErrNotSet, err)
	})
}
