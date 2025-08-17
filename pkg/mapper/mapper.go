package mappers

type MapFunc[T any, U any] func(T) U

func (m MapFunc[T, U]) Map(v T) U {
	return m(v)
}

func (m MapFunc[T, U]) MapEach(v []T) []U {
	result := make([]U, len(v))
	for i, item := range v {
		result[i] = m(item)
	}
	return result
}
