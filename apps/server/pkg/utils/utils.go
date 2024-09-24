package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"unsafe"
)

// B2S converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func B2S(b []byte) string {
	/* #nosec G103 */
	return *(*string)(unsafe.Pointer(&b))
}

// S2B converts a string to a byte slice without memory allocation.
// Note: This method uses unsafe operations and should be used with caution.
func S2B(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

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

func PrettyPrintStruct(data interface{}) {
	prettyJSON, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println("Failed to generate JSON:", err)
		return
	}
	fmt.Println(string(prettyJSON))
}

// Utility function to convert sql.NullInt32 to *int for JSON serialization
func NullIntToPointer(ni sql.NullInt32) *int {
	if ni.Valid {
		val := int(ni.Int32)
		return &val
	}
	return nil
}

// Utility function to convert sql.NullString to *string for JSON serialization
func NullStringToPointer(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func NullBoolToPointer(nb sql.NullBool) *bool {
	if nb.Valid {
		return &nb.Bool
	}
	return nil
}

// Utility function to convert *int to sql.NullInt32
func PointerToNullInt32(ptr *int) sql.NullInt32 {
	if ptr != nil {
		return sql.NullInt32{
			Int32: int32(*ptr),
			Valid: true,
		}
	}
	return sql.NullInt32{
		Int32: 0,
		Valid: false,
	}
}

func Int32ToNullInt32(i int32) sql.NullInt32 {
	if i != 0 {
		return sql.NullInt32{
			Int32: i,
			Valid: true,
		}
	}
	return sql.NullInt32{
		Int32: 0,
		Valid: false,
	}
}

// Utility function to convert *string to sql.NullString
func PointerToNullString(ptr *string) sql.NullString {
	if ptr != nil {
		return sql.NullString{
			String: *ptr,
			Valid:  true,
		}
	}
	return sql.NullString{
		String: "",
		Valid:  false,
	}
}

func StringToNullString(str string) sql.NullString {
	if str != "" {
		return sql.NullString{
			String: str,
			Valid:  true,
		}
	}
	return sql.NullString{
		String: "",
		Valid:  false,
	}
}

// Utility function to convert *bool to sql.NullBool
func PointerToNullBool(ptr *bool) sql.NullBool {
	if ptr != nil {
		return sql.NullBool{
			Bool:  *ptr,
			Valid: true,
		}
	}
	return sql.NullBool{
		Bool:  false,
		Valid: false,
	}
}
