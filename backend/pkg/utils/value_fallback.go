// Package utils provides small shared helpers for common value handling.
package utils

// Default returns fallback when value is the zero value for T.
func Default[T comparable](value, fallback T) T {
	var zero T
	if value != zero {
		return value
	}
	return fallback
}

// DefaultString returns fallback when value is empty.
func DefaultString(value, fallback string) string {
	return Default(value, fallback)
}
