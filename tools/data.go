package tools

import "encoding/json"

// Mapper defines a function signature for mapping a value of type T to type V.
type Mapper[T any, V any] func(t T) V

// Map transforms a slice of type T into a slice of type V using the provided mapper function.
func Map[T any, V any](arr []T, mapper Mapper[T, V]) []V {
	result := make([]V, len(arr))
	for i, item := range arr {
		result[i] = mapper(item)
	}
	return result
}

// ValueTransformer defines a function signature for transforming a value during map traversal.
type ValueTransformer func(val any, path *[]string) any

// ReplaceValue recursively searches for a key in a map (including nested maps and slices of maps)
// and applies the transformer function f to its value. The current traversal path is passed to f.
func ReplaceValue(dest *map[string]any, key string, f ValueTransformer, path *[]string) {
	for k, v := range *dest {
		if k == key {
			(*dest)[k] = f(v, path)
			continue
		}

		switch child := v.(type) {
		case map[string]any:
			updatedPath := append(*path, k)
			ReplaceValue(&child, key, f, &updatedPath)
		case []any:
			updatedPath := append(*path, k)
			for _, element := range child {
				switch ch := element.(type) {
				case map[string]any:
					ReplaceValue(&ch, key, f, &updatedPath)
				}
			}
		default:
		}
	}
}

// Restructure converts data from one type to another using JSON marshaling/unmarshaling.
// This is useful for deep copying or converting between similar structures.
func Restructure[T any](data any, out *T) error {
	h, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(h, out)
}

// RestructureArrays redistributes a 2D slice into a new 2D slice with a fixed number of rows (factor).
// If factor is greater than or equal to input length, it cycles through input rows.
// Otherwise, it distributes input rows by appending them to result rows.
func RestructureArrays(values [][]any, factor int) [][]any {
	if values == nil || len(values) == 0 {
		return nil
	}

	result := make([][]any, factor)

	if factor >= len(values) {
		j := 0
		for i := 0; i < factor; i++ {
			result[i] = values[j]
			j++
			if j == len(values) {
				j = 0
			}
		}
		return result
	}
	i := 0
	for _, v := range values {

		if result[i] == nil {
			result[i] = v
			continue
		}
		result[i] = append(result[i], v...)
		i++
		if i == factor {
			i = 0
		}
	}
	return result

}
