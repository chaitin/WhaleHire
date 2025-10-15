package util

import (
	"context"
	"sync"
)

// 通用并发执行器
type TaskExecutor[T any, R any] struct {
	MaxWorkers int
	TaskFunc   func(context.Context, T) (R, error)
	OnResult   func(T, R, error)
}

func (e *TaskExecutor[T, R]) Run(ctx context.Context, tasks []T) {
	sem := make(chan struct{}, e.MaxWorkers)
	var wg sync.WaitGroup

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			return
		default:
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(t T) {
			defer wg.Done()
			defer func() { <-sem }()

			res, err := e.TaskFunc(ctx, t)
			if e.OnResult != nil {
				e.OnResult(t, res, err)
			}
		}(task)
	}

	wg.Wait()
}
