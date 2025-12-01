package utils

// Ternary returns whenTrue if condition is true, otherwise whenFalse
func Ternary[T any](condition bool, whenTrue T, whenFalse T) T {
	if condition {
		return whenTrue
	}
	return whenFalse
}
