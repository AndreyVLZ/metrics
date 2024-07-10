package flag

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/AndreyVLZ/metrics/pkg/parser/perr"
)

var fs *flag.FlagSet

// init Инициализация глобальных переменных.
func init() {
	var once sync.Once

	once.Do(
		func() {
			fs = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		},
	)
}

func Parse(args []string) error {
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flag: %w", err)
	}

	return nil
}

// Int Читает флаг как int.Возвращает функцию установки int-значения.
func Int(flagName, usage string) func(*int) error {
	var (
		isExistFlag bool
		valInt1     int
	)

	fs.Func(flagName, usage, func(strFlag string) error {
		var err error

		valInt1, err = strconv.Atoi(strFlag)
		if err != nil {
			return fmt.Errorf("fInt conver [%s] to int: %w", strFlag, err)
		}

		isExistFlag = true

		return nil
	})

	return func(valInt *int) error {
		if !isExistFlag {
			return perr.ErrNotSet
		}

		*valInt = valInt1

		return nil
	}
}

// Bool Читает флаг как bool. Возвращает функцию установки bool-значения.
func Bool(flagName, usage string) func(*bool) error {
	var (
		isExistFlag bool
		valBool     bool
	)

	fs.Func(flagName, usage, func(strFlag string) error {
		var err error

		valBool, err = strconv.ParseBool(strFlag)
		if err != nil {
			return fmt.Errorf("fBool convert [%s] to bool: %w", strFlag, err)
		}

		isExistFlag = true

		return nil
	})

	return func(val *bool) error {
		if !isExistFlag {
			return perr.ErrNotSet
		}

		*val = valBool

		return nil
	}
}

// String Читает флаг как string. Возвращает функцию установки string-значения.
func String(flagName, usage string) func(*string) error {
	var (
		isExistFlag bool
		valStr      string
	)

	fs.Func(flagName, usage, func(strFlag string) error {
		isExistFlag = true
		valStr = strFlag

		return nil
	})

	return func(str *string) error {
		if !isExistFlag {
			return perr.ErrNotSet
		}

		*str = valStr

		return nil
	}
}
