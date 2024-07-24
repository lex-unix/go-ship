package lazyloader

import "sync"

type loader[T any] func() (T, error)

type Loader[T any] struct {
	once   *sync.Once
	loader loader[T]
	data   T
	err    error
}

func New[T any](loader loader[T]) *Loader[T] {
	return &Loader[T]{
		once:   &sync.Once{},
		loader: loader,
		err:    nil,
	}
}

func (l *Loader[T]) Load() T {
	l.once.Do(func() {
		data, err := l.loader()
		l.data = data
		l.err = err
	})

	return l.data
}

func (l *Loader[T]) Error() error {
	if l.err != nil {
		return l.err
	}
	return nil
}
