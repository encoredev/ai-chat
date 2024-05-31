package fns

func Ptr[T any](v T) *T {
	return &v
}

func ToMap[K comparable, T any](s []T, f func(T) K) map[K]T {
	m := make(map[K]T, len(s))
	for _, v := range s {
		m[f(v)] = v
	}
	return m
}

func Map[T, U any](s []T, f func(T) U) []U {
	r := make([]U, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}
