package tools

import (
	"reflect"
	"strconv"
	"testing"
)

func TestMap(t *testing.T) {
	t.Run("int to string", func(t *testing.T) {
		input := []int{1, 2, 3}
		want := []string{"1", "2", "3"}
		got := Map(input, func(i int) string {
			return strconv.Itoa(i)
		})
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Map() = %v, want %v", got, want)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		input := []int{}
		got := Map(input, func(i int) int {
			return i * 2
		})
		if len(got) != 0 {
			t.Errorf("Map() length = %v, want 0", len(got))
		}
	})
}

func TestReplaceValue(t *testing.T) {
	t.Run("simple replacement", func(t *testing.T) {
		data := map[string]any{
			"key1": "val1",
			"key2": "val2",
		}
		path := []string{}
		ReplaceValue(&data, "key1", func(val any, p *[]string) any {
			return "replaced"
		}, &path)

		if data["key1"] != "replaced" {
			t.Errorf("ReplaceValue() simple replacement failed, got %v", data["key1"])
		}
	})

	t.Run("nested replacement and path check", func(t *testing.T) {
		data := map[string]any{
			"root": map[string]any{
				"child": "original",
			},
		}
		var capturedPath []string
		path := []string{}
		ReplaceValue(&data, "child", func(val any, p *[]string) any {
			capturedPath = make([]string, len(*p))
			copy(capturedPath, *p)
			return "new"
		}, &path)

		nested := data["root"].(map[string]any)
		if nested["child"] != "new" {
			t.Errorf("ReplaceValue() nested replacement failed")
		}
		wantPath := []string{"root"}
		if !reflect.DeepEqual(capturedPath, wantPath) {
			t.Errorf("ReplaceValue() captured path = %v, want %v", capturedPath, wantPath)
		}
	})

	t.Run("array of maps replacement", func(t *testing.T) {
		data := map[string]any{
			"list": []any{
				map[string]any{"id": 1, "target": "a"},
				map[string]any{"id": 2, "target": "b"},
			},
		}
		path := []string{}
		ReplaceValue(&data, "target", func(val any, p *[]string) any {
			return "X"
		}, &path)

		list := data["list"].([]any)
		for i, item := range list {
			m := item.(map[string]any)
			if m["target"] != "X" {
				t.Errorf("ReplaceValue() item %d failed to replace, got %v", i, m["target"])
			}
		}
	})
}

func TestRestructure(t *testing.T) {
	type Source struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	type Dest struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	t.Run("struct to struct", func(t *testing.T) {
		src := Source{Name: "Alice", Age: 30}
		var dst Dest
		err := Restructure(src, &dst)
		if err != nil {
			t.Fatalf("Restructure() error: %v", err)
		}
		if dst.Name != "Alice" || dst.Age != 30 {
			t.Errorf("Restructure() failed, got %+v", dst)
		}
	})

	t.Run("map to struct", func(t *testing.T) {
		src := map[string]any{"name": "Bob", "age": 25}
		var dst Dest
		err := Restructure(src, &dst)
		if err != nil {
			t.Fatalf("Restructure() error: %v", err)
		}
		if dst.Name != "Bob" || dst.Age != 25 {
			t.Errorf("Restructure() failed, got %+v", dst)
		}
	})
}

func TestRestructureArrays(t *testing.T) {
	t.Run("nil/empty input", func(t *testing.T) {
		if RestructureArrays(nil, 5) != nil {
			t.Error("RestructureArrays(nil) should be nil")
		}
		if RestructureArrays([][]any{}, 5) != nil {
			t.Error("RestructureArrays(empty) should be nil")
		}
	})

	t.Run("factor >= len(values)", func(t *testing.T) {
		values := [][]any{
			{1},
			{2},
		}
		// factor = 3. Expected result: [[1], [2], [1]] (cycles)
		got := RestructureArrays(values, 3)
		want := [][]any{
			{1},
			{2},
			{1},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("RestructureArrays() = %v, want %v", got, want)
		}
	})

	t.Run("factor < len(values)", func(t *testing.T) {
		values := [][]any{
			{1},
			{2},
			{3},
			{4},
		}
		// factor = 2.
		// v1: res[0]=[1], i=0
		// v2: res[0]=[1,2], i=1
		// v3: res[1]=[3], i=1
		// v4: res[1]=[3,4], i=0
		got := RestructureArrays(values, 2)
		want := [][]any{
			{1, 2},
			{3, 4},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("RestructureArrays() = %v, want %v", got, want)
		}
	})

	t.Run("factor < len(values) odd elements", func(t *testing.T) {
		values := [][]any{
			{1},
			{2},
			{3},
		}
		// factor = 2.
		// v1: res[0]=[1], i=0
		// v2: res[0]=[1,2], i=1
		// v3: res[1]=[3], i=1
		got := RestructureArrays(values, 2)
		want := [][]any{
			{1, 2},
			{3},
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("RestructureArrays() = %v, want %v", got, want)
		}
	})
}
