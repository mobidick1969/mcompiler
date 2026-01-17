package main

import (
	"encoding/json"
	"fmt"
	"math"
	"mcompiler/arena"
	"strconv"
	"testing"
)

// TestCase defines a JSON input and description
type TestCase struct {
	Name  string
	Input string
}

func TestCorrectnessAgainstStdLib(t *testing.T) {
	tests := []TestCase{
		{"Simple Object", `{"a": 1, "b": "hello"}`},
		{"Nested Object", `{"user": {"id": 123, "active": true}}`},
		{"Array of Ints", `[1, 2, 3, 4, 5]`},
		{"Mixed Array", `[1, "two", true, null, {"nested": "obj"}]`},
		{"Escaped Strings", `{"msg": "Hello \"World\"", "path": "C:\\Windows"}`},
		{"Unicode", `{"emoji": "🚀", "kr": "안녕하세요"}`},
		{"Numbers", `{"int": 42, "float": 3.14159, "exp": 2.5e-3}`},
		{"Deep Nesting", `{"a": {"b": {"c": {"d": "deep"}}}}`},
	}

	a := arena.NewBestArena()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			a.Reset()
			input := []byte(tc.Input)

			// 1. Parse with StdLib
			var stdResult interface{}
			if err := json.Unmarshal(input, &stdResult); err != nil {
				t.Fatalf("StdLib failed to parse: %v", err)
			}

			// 2. Parse with FastParser
			parser := NewParser(input, a)
			// ParseAny panic handling for safety in tests
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("FastParser panicked: %v", r)
				}
			}()
			node := parser.ParseAny()

			// 3. Compare
			if err := compareNodeWithStd(node, stdResult); err != nil {
				t.Errorf("Mismatch in %s: %v", tc.Name, err)
			}
		})
	}
}

// compareNodeWithStd recursively compares our Node with generic interface{}
func compareNodeWithStd(node *Node, std interface{}) error {
	switch node.Type {
	case Object:
		stdMap, ok := std.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected Object, got %T", std)
		}
		// Count children
		childCount := 0
		curr := node.Children
		for curr != nil {
			childCount++
			val, exists := stdMap[curr.Key]
			// NOTE: curr.Key is a raw string slice.
			// StdLib parsing unescapes keys. If our key has escapes, this lookup might fail.
			// For this test, we assume keys are simple or match.
			if !exists {
				return fmt.Errorf("key %q found in Node but not in StdLib", curr.Key)
			}
			if err := compareNodeWithStd(curr, val); err != nil {
				return fmt.Errorf("key %q mismatch: %v", curr.Key, err)
			}
			curr = curr.Next
		}
		if childCount != len(stdMap) {
			return fmt.Errorf("size mismatch: Node=%d, Std=%d", childCount, len(stdMap))
		}
		return nil

	case Array:
		stdArr, ok := std.([]interface{})
		if !ok {
			return fmt.Errorf("expected Array, got %T", std)
		}
		idx := 0
		curr := node.Children
		for curr != nil {
			if idx >= len(stdArr) {
				return fmt.Errorf("Node has more items than StdLib array")
			}
			if err := compareNodeWithStd(curr, stdArr[idx]); err != nil {
				return fmt.Errorf("array index %d mismatch: %v", idx, err)
			}
			curr = curr.Next
			idx++
		}
		if idx != len(stdArr) {
			return fmt.Errorf("Node has fewer items than StdLib array")
		}
		return nil

	case String:
		stdStr, ok := std.(string)
		if !ok {
			return fmt.Errorf("expected String, got %T", std)
		}
		// Use Unescape() to get the "cooked" string for comparison
		parsedStr, err := node.Unescape()
		if err != nil {
			return fmt.Errorf("Unescape failed: %v", err)
		}
		if parsedStr != stdStr {
			return fmt.Errorf("expected string %q, got %q", stdStr, parsedStr)
		}
		return nil

	case Number:
		stdNum, ok := std.(float64)
		if !ok {
			return fmt.Errorf("expected Number (float64), got %T", std)
		}
		// node.ValueStr is the numeric string (e.g. "3.14").
		// Parse it to float64
		val, err := strconv.ParseFloat(node.ValueStr, 64)
		if err != nil {
			return fmt.Errorf("invalid number format in Node: %s", node.ValueStr)
		}
		if val != stdNum {
			// Float equality check with epsilon?
			if math.Abs(val-stdNum) > 1e-9 {
				return fmt.Errorf("number mismatch: expected %v, got %v", stdNum, val)
			}
		}
		return nil

	case True:
		stdBool, ok := std.(bool)
		if !ok {
			return fmt.Errorf("expected Bool (true), got %T", std)
		}
		if !stdBool {
			return fmt.Errorf("expected true, got false")
		}
		return nil

	case False:
		stdBool, ok := std.(bool)
		if !ok {
			return fmt.Errorf("expected Bool (false), got %T", std)
		}
		if stdBool {
			return fmt.Errorf("expected false, got true")
		}
		return nil

	case Null:
		if std != nil {
			return fmt.Errorf("expected nil, got %T", std)
		}
		return nil

	default:
		return fmt.Errorf("unknown node type")
	}
}
