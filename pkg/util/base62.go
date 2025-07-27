
// 文件路径: pkg/util/base62.go
package util

import (
	"strings"
)

const (
	base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base        = uint64(len(base62Chars))
)

// ToBase62 将一个uint64整数编码为Base62字符串
func ToBase62(n uint64) string {
	if n == 0 {
		return string(base62Chars[0])
	}

	var sb strings.Builder
	for n > 0 {
		r := n % base
		n /= base
		sb.WriteByte(base62Chars[r])
	}

	// 反转字符串
	runes := []rune(sb.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}