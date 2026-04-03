package utils

import (
	ulid "github.com/oklog/ulid/v2"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// utils/string.go
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
func ToPtr[T any](v T) *T {
	return &v
}
func GenerateDatabaseId() string {
	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	ms := ulid.Timestamp(time.Now())
	id, _ := ulid.New(ms, entropy)
	m := strings.ToLower(id.String())
	return m
}
func UintArrayEqual(arr1, arr2 []uint) bool {
	arr1Map := make(map[uint]uint)
	arr2Map := make(map[uint]uint)
	for _, item := range arr1 {
		arr1Map[item] = item
	}
	for _, item := range arr2 {
		arr2Map[item] = item
	}
	if len(arr1Map) != len(arr2Map) {
		return false
	}
	for k, _ := range arr1Map {
		if _, exist := arr2Map[k]; !exist {
			return false
		}
	}
	return true
}

func SafeNestedString(obj map[string]interface{}, fields ...string) (string, bool) {
	var current interface{} = obj
	for _, field := range fields {
		if currentMap, ok := current.(map[string]interface{}); ok {
			current = currentMap[field]
			if current == nil {
				return "", false
			}
		} else if currentSlice, ok := current.([]interface{}); ok {
			// 处理数组索引（field 应为数字字符串，如 "0"）
			if idx, err := strconv.Atoi(field); err == nil && idx >= 0 && idx < len(currentSlice) {
				current = currentSlice[idx]
			} else {
				return "", false
			}
		} else {
			return "", false
		}
	}
	if str, ok := current.(string); ok {
		return str, true
	}
	return "", false
}
