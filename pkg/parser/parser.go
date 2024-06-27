package parser

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/AndreyVLZ/metrics/pkg/parser/field"
	"github.com/AndreyVLZ/metrics/pkg/parser/flag"
	"github.com/AndreyVLZ/metrics/pkg/parser/perr"
)

type valPtr interface {
	*int | *string | *bool | *time.Duration
}

var myParser *parser

type parser struct {
	listFn []func() error
}

// init Инициализация глобальных переменных.
func init() {
	var once sync.Once

	once.Do(
		func() {
			myParser = &parser{
				listFn: make([]func() error, 0),
			}
		},
	)
}

func (p *parser) parse(args []string) error {
	if err := flag.Parse(args); err != nil {
		return fmt.Errorf("parse flag: %w", err)
	}

	for _, fnSet := range p.listFn {
		if err := fnSet(); err != nil {
			return err
		}
	}

	return nil
}

func (p *parser) addFn(fn func() error) { p.listFn = append(p.listFn, fn) }

// Parse ...
func Parse(args []string) error { return myParser.parse(args) }

// Value ...
func Value[T valPtr](defVal T, parsers ...func(T) error) {
	myParser.addFn(
		func() error {
			for i := range parsers {
				if err := parsers[i](defVal); err != nil {
					if !errors.Is(err, perr.ErrNotSet) {
						return err
					}
				}
			}

			return nil
		})
}

// File ...
func File(defVal *string, parsers ...func(*string) error) {
	myParser.addFn(
		func() error {
			for i := range parsers {
				if err := parsers[i](defVal); err != nil {
					if !errors.Is(err, perr.ErrNotSet) {
						return err
					}
				}
			}

			if *defVal == "" {
				return nil
			}

			fileByte, err := os.ReadFile(*defVal)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			if err := field.Unmarshal(fileByte); err != nil {
				return err
			}

			return nil
		})
}
