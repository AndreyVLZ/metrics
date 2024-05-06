// Сохраняет в пямати метрики и возвращает сохраненные значения по имени
package cache

import (
	_ "net/http/pprof"
	"sync"

	"github.com/AndreyVLZ/metrics/internal/model"
)

type Cache[VT model.ValueType] struct {
	mu   sync.Mutex
	repo map[string]*model.MetricRepo[VT]
}

func New[VT model.ValueType]() *Cache[VT] {
	return &Cache[VT]{
		mu:   sync.Mutex{},
		repo: make(map[string]*model.MetricRepo[VT], 0),
	}
}

// Получение списка метрик.
func (c *Cache[VT]) List() []model.MetricRepo[VT] {
	c.mu.Lock()
	defer c.mu.Unlock()

	arr := make([]model.MetricRepo[VT], 0, len(c.repo))
	for _, met := range c.repo {
		arr = append(arr, *met)
	}

	return arr
}

// Установка новой метрики.
func (c *Cache[VT]) Set(met model.MetricRepo[VT]) model.MetricRepo[VT] {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.repo[met.Name()] = &met

	return met
}

// Получение метрики по имени.
func (c *Cache[VT]) Get(name string) (model.MetricRepo[VT], bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	mDB, ok := c.repo[name]
	if !ok {
		return model.MetricRepo[VT]{}, false
	}

	return *mDB, true
}
