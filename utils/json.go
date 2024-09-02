package utils

import (
	jsoniter "github.com/json-iterator/go"
)

// 定义JSON操作
var (
	json             = jsoniter.ConfigCompatibleWithStandardLibrary
	Marshal          = json.Marshal
	UnmarshalX       = json.Unmarshal
	MarshalIndent    = json.MarshalIndent
	MarshalToStringX = json.MarshalToString
	NewDecoder       = json.NewDecoder
	NewEncoder       = json.NewEncoder
)

// MarshalToString JSON编码为字符串
func MarshalToString(v any) string {
	s, err := MarshalToStringX(v)
	if err != nil {
		return ""
	}
	return s
}

func Unmarshal(v []byte, o any) error {
	return UnmarshalX(v, o)
}
