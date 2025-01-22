package hook

import (
	"crypto/md5"
	"encoding/hex"
	"unsafe"
)

// MD5 计算 md5
func MD5(s string) string {
	b := md5.Sum(unsafe.Slice(unsafe.StringData(s), len(s)))
	return hex.EncodeToString(b[:])
}
