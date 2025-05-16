package lazy

import "sync"

type LoadFunc[T any] func() (T, error)

type Lazy[T any] struct {
	once   *sync.Once
	loader LoadFunc[T]
	data   T
	err    error
}

func New[T any](loader LoadFunc[T]) *Lazy[T] {
	return &Lazy[T]{
		once:   &sync.Once{},
		loader: loader,
		err:    nil,
	}
}

func (l *Lazy[T]) Load() (T, error) {
	l.once.Do(func() {
		data, err := l.loader()
		l.data = data
		l.err = err
	})

	return l.data, l.err
}

func (l *Lazy[T]) Error() error {
	if l.err != nil {
		return l.err
	}
	return nil
}
