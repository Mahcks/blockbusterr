package utils

import "reflect"

// IsEmptyValue uses reflection to determine if a value is empty.
func IsEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// Util - Ternary:
// A golang equivalent to JS Ternary Operator
//
// It takes a condition, and returns a result depending on the outcome
func Ternary[T any](condition bool, whenTrue T, whenFalse T) T {
	if condition {
		return whenTrue
	}

	return whenFalse
}
