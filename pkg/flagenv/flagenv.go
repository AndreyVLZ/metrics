package flagenv

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
)

type FnSet func() error

func New(setters ...FnSet) error {
	errs := make([]error, 0, len(setters))

	for i := range setters {
		if err := setters[i](); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func String(addr *string, envName, flagName, defValue, usage string) FnSet {
	return func() error {
		flag.StringVar(addr, flagName, orEnvString(envName, defValue), usage)

		return nil
	}
}

func orEnvString(envName, defValue string) string {
	if val, ok := os.LookupEnv(envName); ok {
		return val
	}

	return defValue
}

func Bool(boolPtr *bool, envName, flagName string, defValue bool, usage string) FnSet {
	return func() error {
		flag.BoolVar(boolPtr, flagName, orEnvBool(envName, defValue), usage)

		return nil
	}
}

func orEnvBool(envName string, defValue bool) bool {
	if valStr, ok := os.LookupEnv(envName); ok {
		return valStr == "true"
	}

	return defValue
}

func Int(intPtr *int, envName, flagName string, defValue int, usage string) FnSet {
	return func() error {
		var err error

		defValue, err = orEnvInt(envName, defValue)
		if err != nil {
			return fmt.Errorf("parse env [%s]: %w", envName, err)
		}

		flag.IntVar(intPtr, flagName, defValue, usage)

		return nil
	}
}

func orEnvInt(envName string, defValue int) (int, error) {
	if valStr, ok := os.LookupEnv(envName); ok {
		intVal, err := strconv.Atoi(valStr)
		if err != nil {
			return 0, fmt.Errorf("to int: %w", err)
		}

		return intVal, nil
	}

	return defValue, nil
}
