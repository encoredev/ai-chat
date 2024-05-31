package fns

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}

// ToMap converts a slice to a map using the given key function.
func ToMap[K comparable, T any](s []T, f func(T) K) map[K]T {
	m := make(map[K]T, len(s))
	for _, v := range s {
		m[f(v)] = v
	}
	return m
}

// Map applies the given function to each element in the slice and returns a new slice with the results.
func Map[T, U any](s []T, f func(T) U) []U {
	r := make([]U, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}
