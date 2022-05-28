package util

import (
	"encoding/json"

	"github.com/google/uuid"
)

func UuidString() string {
	return uuid.New().String()
}

func ToJson(v interface{}) string {
	result, _ := json.Marshal(v)
	return string(result)
}
