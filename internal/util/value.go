package util

func NewValue[T comparable](v T) *T {
	return &v
}