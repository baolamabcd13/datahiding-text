package utils

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateRandomString - Tạo chuỗi ngẫu nhiên với độ dài xác định
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
} 