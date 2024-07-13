// Реализация задачи и пула задач.
package task

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Task Структура включает в себя:
// наименование для задачи,
// функцию-работу для задачи,
// задержку между выполениями работы.
type Task struct {
	fn       func() error
	name     string
	duration time.Duration
}

// New Возвращает новую задачу.
func New(name string, duration time.Duration, fn func() error) Task {
	return Task{
		fn:       fn,
		name:     name,
		duration: duration,
	}
}

// Poll Структура включает в себя срез задач и логер.
type Poll struct {
	log   *slog.Logger
	tasks []Task
}

// NewPoll Возвращает новый пул задач.
// [size] - начальная вместимость для среза задач.
func NewPoll(size int, log *slog.Logger) *Poll {
	return &Poll{
		tasks: make([]Task, 0, size),
		log:   log,
	}
}

// Add Добавляет новый срез задач в пул.
func (p *Poll) Add(tasks ...Task) { p.tasks = append(p.tasks, tasks...) }

// Run Запуск пула.
// Возвращает канал [chErr] с ошибками при выполнении задач.
// В фоне ожидает завершение работы всех задач.
// Канал [chErr] закрывается когда все задачи завершены.
// Отмена контекста [ctx] отменяет все задачи.
func (p *Poll) Run(ctx context.Context) <-chan error {
	var wgTask sync.WaitGroup

	chErr := make(chan error)

	for idTask := range p.tasks {
		wgTask.Add(1)

		go func(task Task) {
			p.log.Debug("запуск задачи", "name", task.name)

			defer wgTask.Done()

			for {
				select {
				case <-ctx.Done():
					p.log.Debug("отмена задачи", "name", task.name)

					return
				case <-time.After(task.duration):
					if err := task.fn(); err != nil {
						chErr <- err
					}
				}
			}
		}(p.tasks[idTask])
	}

	go func() {
		wgTask.Wait() // Ждем завершения работы всех задач [Task]
		close(chErr)  // Закрываем канал в который писали Task'и
		p.log.Debug("task poll", "wait", "ok")
	}()

	return chErr
}
